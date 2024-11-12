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
	openaiClient := openai.NewClient(openaiCreds.Token)

	for {
		err := mainLoop(ctx, redditClient, openaiClient)
		if err != nil {
			fmt.Println(err)
		}
	}

}

func mainLoop(ctx context.Context,
	redditClient *reddit.Client, openaiClient *openai.Client,
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
	} else {

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
		Model: "gpt-3.5-turbo",
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleAssistant, Content: automodComment.Body},
			{Role: openai.ChatMessageRoleUser, Content: makeUserContext(post, opReply)},
		},
		Functions: modFunctions,
	})
	if err != nil {
		return fmt.Errorf("failed to create chat completion: %w", err)
	}

	switch resp.Choices[0].Message.FunctionCall.Name {
	case "approve":
		approval, e := UnmarshalApproval([]byte(resp.Choices[0].Message.FunctionCall.Arguments))
		if e != nil {
			return fmt.Errorf("failed to unmarshal approval: %w", e)
		}
		err = suggestApprove(approval)

	case "remove":
		removal, e := UnmarshalRemoval([]byte(resp.Choices[0].Message.FunctionCall.Arguments))
		if e != nil {
			return fmt.Errorf("failed to unmarshal removal: %w", e)
		}
		err = suggestRemove(removal)
	}
	if err != nil {
		return err
	}

	_, err = fmt.Printf("You can \"%s%s\", \"%s%s\" %s or %s %s: ",

		aurora.Underline("a").Green(),
		aurora.Green("pprove"),

		aurora.Underline("r").Red(),
		aurora.Red("emove"),
		aurora.Gray(12, "(no removal reason)"),

		aurora.Gray(12, "skip"),
		aurora.Gray(12, "(default)"),
	)

	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	switch strings.TrimSpace(input) {
	case "a", "approve":
		_, err := redditClient.Moderation.Approve(ctx, post.Post.FullID)
		if err != nil {
			return fmt.Errorf("failed to approve post: %w", err)
		}

		_, err = fmt.Println("Approved!")
		if err != nil {
			return fmt.Errorf("failed to print approval message: %w", err)
		}

	case "r", "remove":
		_, err := redditClient.Moderation.Remove(ctx, post.Post.FullID)
		if err != nil {
			return fmt.Errorf("failed to remove post: %w", err)
		}

		_, err = fmt.Println("Removed!")
		if err != nil {
			return fmt.Errorf("failed to print removal message: %w", err)
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

func suggestApprove(approval Approval) error {
	_, err := fmt.Printf(`Recommendation: %s
Someone: %s
Something: %s
Consequences: %s
Explanation: %s

`,
		aurora.Green("Approve"),
		aurora.Gray(6, approval.Someone),
		aurora.Gray(6, approval.Something),
		aurora.Gray(6, approval.Consequences),
		aurora.Gray(12, approval.Explanation),
	)

	if err != nil {
		return fmt.Errorf("failed to suggest approval of a post: %w", err)
	}

	return err
}

func suggestRemove(removal Removal) error {
	_, err := fmt.Printf(`Recommendation: %s
Reason: %s

`, aurora.Red("Remove"), removal.Reason)

	if err != nil {
		return fmt.Errorf("failed to suggest removal of a post: %w", err)
	}

	return err
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
