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

	_, err = fmt.Printf("Logged in as %s\n",
		aurora.Red("u/"+user.Name).Hyperlink("https://reddit.com/u/"+user.Name),
	)
	if err != nil {
		panic(err)
	}

	openaiCreds, err := readOpenAiCredentials(dir)
	if err != nil {
		panic(err)
	}
	config := openai.DefaultConfig(openaiCreds.Token)
	config.BaseURL = openaiCreds.BaseURL
	openaiClient := openai.NewClientWithConfig(config)

	for {
		err := mainLoop(ctx, redditClient, openaiClient, openaiCreds.Model)
		if err != nil {
			fmt.Println(err)
		}
	}

}

func mainLoop(ctx context.Context,
	redditClient *reddit.Client,
	openaiClient *openai.Client, model string,
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

	_, err = fmt.Printf(`
Title: %s
Body: %s
URL: %s

`,
		aurora.Bold(post.Post.Title).Hyperlink(PermaLink(post.Post.Permalink)),
		aurora.Gray(12, post.Post.Body),
		aurora.Italic(post.Post.URL),
	)
	if err != nil {
		return fmt.Errorf("failed to print post: %w", err)
	}

	automodComment, err := FindAutomodComment(post)
	if err != nil {
		return fmt.Errorf("failed to find AutoModerator comment: %w", err)
	}

	_, err = fmt.Printf("Found %s by %s\n",
		aurora.Hyperlink("comment", PermaLink(automodComment.Permalink)),
		aurora.Green("u/"+automodComment.Author).Hyperlink("https://reddit.com/u/"+automodComment.Author),
	)
	if err != nil {
		return fmt.Errorf("failed to print u/AutoModerator's comment: %w", err)
	}

	_, err = redditClient.Comment.LoadMoreReplies(ctx, automodComment)
	if err != nil {
		return fmt.Errorf("failed to load more replies: %w", err)
	}

	opReply, err := FindExplanatoryComment(post, automodComment)
	if err != nil {
		fmt.Printf("Failed to find explanatory comment: %v\n", err)
	}

	if opReply != nil {

		_, err = fmt.Printf(`Found %s by %s
Body: %s

`,
			aurora.Hyperlink("explanatory comment", PermaLink(opReply.Permalink)),
			aurora.Red("u/"+opReply.Author).Hyperlink("https://reddit.com/u/"+opReply.Author),
			aurora.Gray(12, opReply.Body),
		)
		if err != nil {
			return fmt.Errorf("failed to print explanatory comment: %w", err)
		}
	}

	resp, err := openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleAssistant, Content: automodComment.Body},
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
	_, err = fmt.Printf("You can \"%s%s\", \"%s%s\" %s or %s %s: ",

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
		fmt.Println("Removing...")
		_, err := redditClient.Moderation.Remove(ctx, post.Post.FullID)
		if err != nil {
			return fmt.Errorf("failed to remove post: %w", err)
		}

		if removal != nil {
			removalMessage, err := formatRemovalMessage(removal.Reason)
			if err != nil {
				return fmt.Errorf("failed to format removal message: %w", err)
			}

			fmt.Println("Adding removal reason...")
			removalComment, _, err := redditClient.Comment.Submit(ctx, post.Post.FullID, removalMessage)
			if err != nil {
				return fmt.Errorf("failed to submit removal reason: %w", err)
			}

			fmt.Println("Distinguishing and stickying removal reason...")
			_, err = redditClient.Moderation.DistinguishAndSticky(ctx, removalComment.ID)
			if err != nil {
				return fmt.Errorf("failed to distinguish and sticky removal reason: %w", err)
			}
		}

		fmt.Println("Removed!")

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

func reasonOrNone(removal *Removal) string {
	if removal == nil {
		return "(no removal reason)"
	}

	return fmt.Sprintf("(%s)", removal.Reason)
}

func formatExplanatoryComment(opReply *reddit.Comment) string {
	if opReply == nil {
		return ""
	}

	return strings.Join(strings.Split(opReply.Body, "\n"), "\t")
}

func makeUserContext(post *reddit.PostAndComments, opReply *reddit.Comment) string {
	postBody := strings.Join(strings.Split(post.Post.Body, "\n"), "\t")
	commentBody := formatExplanatoryComment(opReply)
	return fmt.Sprintf(`Post title: %s
Post body: %s
Explanatory comment: %s`, post.Post.Title, postBody, commentBody)
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
