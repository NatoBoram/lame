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
		err := handleEntrypoint(ctx, redditClient, guide, openaiClient, openaiCreds.Model)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func handleEntrypoint(ctx context.Context,
	redditClient *reddit.Client, guide *reddit.PostAndComments,
	openaiClient *openai.Client, model string,
) error {
	fmt.Printf("Enter a Reddit post url or the name of a feed (%s%s, %s%s, %s%s, %s%s): ",
		aurora.Gray(12, "h").Underline(),
		aurora.Gray(12, "ot"),

		aurora.Gray(12, "n").Underline(),
		aurora.Gray(12, "ew"),

		aurora.Gray(12, "t").Underline(),
		aurora.Gray(12, "op"),

		aurora.Gray(12, "r").Underline(),
		aurora.Gray(12, "ising"),
	)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read post url or feed name: %w", err)
	}

	feed := toRedditFeed(strings.TrimSpace(input))
	if feed == "" {
		err := handlePost(ctx, redditClient, guide, openaiClient, model, input)
		if err != nil {
			return fmt.Errorf("failed to handle post url: %w", err)
		}
	} else {
		err := handleFeed(ctx, redditClient, guide, openaiClient, model, feed)
		if err != nil {
			return fmt.Errorf("failed to handle feed: %w", err)
		}
	}

	return err
}

func handleFeed(ctx context.Context,
	redditClient *reddit.Client, guide *reddit.PostAndComments,
	openaiClient *openai.Client, model string,

	feed RedditFeed,
) error {
	var after string

	for {
		opts := maybeOptions(after)
		posts, _, err := getFeedPosts(ctx, redditClient, feed, opts)
		if err != nil {
			return fmt.Errorf("failed to get feed posts: %w", err)
		}

		err = handlePage(ctx, redditClient, guide, openaiClient, model, posts)
		if err != nil {
			return fmt.Errorf("failed to handle feed page: %w", err)
		}

		after = posts[len(posts)-1].FullID
	}
}

func handlePage(
	ctx context.Context,
	redditClient *reddit.Client, guide *reddit.PostAndComments,
	openaiClient *openai.Client, model string,

	posts []*reddit.Post,
) error {
	for _, post := range posts {
		// TODO: Check for `approved`. go-reddit is missing this field and it's
		// abandoned so I can't even submit a pull request.
		if post.Stickied || post.Locked {
			continue
		}

		err := handlePost(ctx, redditClient, guide, openaiClient, model, post.ID)
		if err != nil {
			return fmt.Errorf("failed to handle post: %w", err)
		}
	}

	return nil
}

func handlePost(ctx context.Context,
	redditClient *reddit.Client, guide *reddit.PostAndComments,
	openaiClient *openai.Client, model string,

	postId string,
) error {
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
	fmt.Printf("You can \"%s%s\", \"%s%s\" %s or %s%s %s: ",

		aurora.Underline("a").Green(),
		aurora.Green("pprove"),

		aurora.Underline("r").Red(),
		aurora.Red("emove"),
		aurora.Gray(12, promptRemoval),

		aurora.Gray(12, "s").Underline(),
		aurora.Gray(12, "kip"),
		aurora.Gray(12, "(default)"),
	)

	reader := bufio.NewReader(os.Stdin)
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

		fmt.Println("Approved!")

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
					fmt.Println("Skipped.")
					fmt.Println()
					return nil
				}
			}
		}

		fmt.Println("Removing...")
		_, err := redditClient.Moderation.Remove(ctx, post.Post.FullID)
		if err != nil {
			return fmt.Errorf("failed to remove post: %w", err)
		}
		fmt.Println("Removed.")

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
			fmt.Println("Done.")
		}

	case "", "s", "skip":
		fmt.Println("Skipped.")

	default:
		fmt.Println("Invalid input. Skipped.")

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
	case "n", "no", "":
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

	fmt.Println(string(s))
	return err
}
