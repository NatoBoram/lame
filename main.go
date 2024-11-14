// Lame is a LLM-powered verification tool for explanatory comments in
// [r/LeopardsAteMyFace].
//
// [r/LeopardsAteMyFace]: https://www.reddit.com/r/LeopardsAteMyFace
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/logrusorgru/aurora/v4"
	openai "github.com/sashabaranov/go-openai"
	"github.com/vartanbeno/go-reddit/v2/reddit"
)

const version = "0.0.0"
const packageName = "github.com/NatoBoram/lame"

func main() {
	dir, err := lameConfigDir()
	if err != nil {
		panic(err)
	}

	redCreds, err := readRedditCredentials(dir)
	if err != nil {
		panic(err)
	}

	redditCredentials := reddit.Credentials{
		ID: redCreds.ID, Secret: redCreds.Secret,
		Username: redCreds.Username, Password: redCreds.Password,
	}

	ua := fmt.Sprintf("%s:%s:%s (by /u/NatoBoram)",
		runtime.GOOS, packageName, version,
	)

	opts := reddit.WithUserAgent(ua)
	redditClient, err := reddit.NewClient(redditCredentials, opts)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	user, _, err := redditClient.Account.Info(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Logged in as %s\n",
		aurora.Red("u/"+user.Name).Hyperlink("https://reddit.com/u/"+user.Name),
	)

	openaiCreds, err := readOpenAiCredentials(dir)
	if err != nil {
		panic(err)
	}
	config := openai.DefaultConfig(openaiCreds.Token)
	config.BaseURL = openaiCreds.BaseURL
	openaiClient := openai.NewClientWithConfig(config)

	// A guide to this sub's explanatory comment rule.
	guide, _, err := redditClient.Post.Get(ctx, redCreds.Guide)
	if err != nil {
		panic(err)
	}

	for {
		err := mainLoop(ctx, redditClient, openaiClient, openaiCreds.Model, guide)
		if err != nil {
			fmt.Println(err)
		}
	}

}

func mainLoop(ctx context.Context,
	redditClient *reddit.Client,
	openaiClient *openai.Client, model string,
	guide *reddit.PostAndComments,
) error {
	_, err := fmt.Print("Enter a Reddit post url: ")
	if err != nil {
		return fmt.Errorf("failed to print prompt: %w", err)
	}

	reader := bufio.NewReader(os.Stdin)
	url, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	postId, err := GetPostId(url)
	if err != nil {
		return fmt.Errorf("failed to get post id: %w", err)
	}

	post, _, err := redditClient.Post.Get(ctx, postId)
	if err != nil {
		return fmt.Errorf("failed to get post: %w", err)
	}

	fmt.Printf(`
Title: %s
Body: %s
URL: %s

`,
		aurora.Bold(post.Post.Title).Hyperlink(PermaLink(post.Post.Permalink)),
		aurora.Gray(12, post.Post.Body),
		aurora.Italic(post.Post.URL),
	)

	automodComment, err := FindAutomodComment(post)
	if err != nil {
		return fmt.Errorf("failed to find AutoModerator comment: %w", err)
	}

	fmt.Printf("Found %s by %s\n",
		aurora.Hyperlink("comment", PermaLink(automodComment.Permalink)),
		aurora.Green("u/"+automodComment.Author).Hyperlink("https://reddit.com/u/"+automodComment.Author),
	)

	_, err = redditClient.Comment.LoadMoreReplies(ctx, automodComment)
	if err != nil {
		return fmt.Errorf("failed to load more replies: %w", err)
	}

	opReply, err := FindExplanatoryComment(post, automodComment)
	if err != nil {
		fmt.Printf("Failed to find explanatory comment: %v\n", err)
	}

	if opReply != nil {
		fmt.Printf(`Found %s by %s
Body: %s

`,
			aurora.Hyperlink("explanatory comment", PermaLink(opReply.Permalink)),
			aurora.Red("u/"+opReply.Author).Hyperlink("https://reddit.com/u/"+opReply.Author),
			aurora.Gray(12, opReply.Body),
		)
	}

	resp, err := openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemMessage},
			{Role: openai.ChatMessageRoleAssistant, Content: guide.Post.Body},
			{Role: openai.ChatMessageRoleUser, Content: makeUserContext(post, opReply)},
		},
		Tools: modTools,
	})
	if err != nil {
		return fmt.Errorf("failed to create chat completion: %w", err)
	}

	_, removal, err := suggest(resp)
	if err != nil {
		return fmt.Errorf("failed to suggest approval or removal: %w", err)
	}

	promptRemoval := reasonOrNone(removal)
	fmt.Printf("You can \"%s%s\", \"%s%s\" %s or %s %s: ",

		aurora.Underline("a").Green(),
		aurora.Green("pprove"),

		aurora.Underline("r").Red(),
		aurora.Red("emove"),
		aurora.Gray(12, promptRemoval),

		aurora.Gray(12, "skip"),
		aurora.Gray(12, "(default)"),
	)

	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	switch strings.TrimSpace(input) {
	case "a", "approve":
		fmt.Println("Approving...")
		_, err := redditClient.Moderation.Approve(ctx, post.Post.FullID)
		if err != nil {
			return fmt.Errorf("failed to approve post: %w", err)
		}

		_, err = fmt.Println("Approved!")
		if err != nil {
			return fmt.Errorf("failed to print approval message: %w", err)
		}

	case "r", "remove":
		if removal == nil {
			fmt.Println("Getting new removal reason...")
			removal, err = retryRemovalReason(post, automodComment, opReply, ctx, model, openaiClient)
			if err != nil {
				fmt.Printf("Failed to get another removal reason: %v\n", err)
			}

			if removal != nil {
				fmt.Printf("Got new removal reason: %s\n", removal.Reason)

				ok, err := confirmNewRemovalReason()
				if err != nil {
					return fmt.Errorf("failed to confirm new removal reason: %w", err)
				}
				if !ok {
					fmt.Println("Skipping...")
					return nil
				}
			}
		}

		fmt.Println("Removing...")
		_, err := redditClient.Moderation.Remove(ctx, post.Post.FullID)
		if err != nil {
			return fmt.Errorf("failed to remove post: %w", err)
		}
		fmt.Println("Removed!")

		if removal != nil {
			removalMessage, err := formatRemovalMessage(removal.Reason, model)
			if err != nil {
				return fmt.Errorf("failed to format removal message: %w", err)
			}

			fmt.Println("Adding removal reason...")
			removalComment, _, err := redditClient.Comment.Submit(ctx, post.Post.FullID, removalMessage)
			if err != nil {
				return fmt.Errorf("failed to submit removal reason: %w", err)
			}
			fmt.Printf("Removal reason added: %s\n", PermaLink(removalComment.Permalink))

			fmt.Println("Distinguishing and stickying removal reason...")
			_, err = redditClient.Moderation.DistinguishAndSticky(ctx, removalComment.FullID)
			if err != nil {
				return fmt.Errorf("failed to distinguish and sticky removal reason: %w", err)
			}
		}

	case "", "s", "skip":
		_, err = fmt.Println("Skipping...")
		if err != nil {
			return fmt.Errorf("failed to print skip message: %w", err)
		}

	default:
		_, err = fmt.Println("Invalid input. Skipping...")
		if err != nil {
			return fmt.Errorf("failed to print invalid input message: %w", err)
		}
	}

	_, err = fmt.Println()
	if err != nil {
		return fmt.Errorf("failed to print newline: %w", err)
	}

	return err
}

func confirmNewRemovalReason() (bool, error) {
	fmt.Printf("Proceed with new removal reason? (%s%s/%s%s): ",
		aurora.Underline("y").Green(),
		aurora.Green("es"),

		aurora.Underline("n").Red(),
		aurora.Red("o"),
	)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read confirmation input: %w", err)
	}

	switch strings.TrimSpace(input) {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		fmt.Println("Invalid input.")
		return false, nil
	}
}

func reasonOrNone(removal *Removal) string {
	if removal == nil {
		return "(will fetch another removal reason)"
	}

	return fmt.Sprintf("(%s)", removal.Reason)
}

func prettyPrint(i interface{}) error {
	s, err := json.MarshalIndent(i, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	_, err = fmt.Println(string(s))
	if err != nil {
		return fmt.Errorf("failed to print JSON: %w", err)
	}

	return err
}
