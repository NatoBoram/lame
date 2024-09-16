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
	fmt.Print("Enter a Reddit post url: ")

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

	redditClient.Comment.LoadMoreReplies(ctx, automodComment)

	opReply, err := FindExplanatoryComment(post, automodComment)
	if err != nil {
		return err
	}

	fmt.Printf(`Found %s by %s
Body: %s

`,
		aurora.Hyperlink("explanatory comment", PermaLink(opReply.Permalink)),
		aurora.Red("u/"+opReply.Author).Hyperlink("https://reddit.com/u/"+opReply.Author),
		aurora.Gray(12, opReply.Body),
	)

	resp, err := openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleAssistant, Content: systemMessage},
			{Role: openai.ChatMessageRoleUser, Content: makeUserContext(post, opReply)},
		},
		Functions: modFunctions,
	})
	if err != nil {
		return fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		prettyPrint(resp)
		return fmt.Errorf("no response from OpenAI")
	}

	if resp.Choices[0].Message.FunctionCall == nil {
		prettyPrint(resp)
		return fmt.Errorf("no function call in response")
	}

	switch resp.Choices[0].Message.FunctionCall.Name {
	case "approve":
		approval, e := UnmarshalApproval([]byte(resp.Choices[0].Message.FunctionCall.Arguments))
		if e != nil {
			return fmt.Errorf("failed to unmarshal approval: %w", e)
		}
		suggestApprove(approval)

	case "remove":
		removal, e := UnmarshalRemoval([]byte(resp.Choices[0].Message.FunctionCall.Arguments))
		if e != nil {
			return fmt.Errorf("failed to unmarshal removal: %w", e)
		}
		suggestRemove(removal)
	}

	fmt.Printf("You can \"%s%s\", \"%s%s\" %s or %s %s: ",

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

		fmt.Println("Approved!")

	case "r", "remove":
		_, err := redditClient.Moderation.Remove(ctx, post.Post.FullID)
		if err != nil {
			return fmt.Errorf("failed to remove post: %w", err)
		}

		fmt.Println("Removed!")

	case "", "s", "skip":
		fmt.Println("Skipping...")

	default:
		fmt.Println("Invalid input. Skipping...")
	}

	fmt.Println()
	return err
}

func makeUserContext(post *reddit.PostAndComments, opReply *reddit.Comment) string {
	message := ""

	if post.Post.Title != "" {
		message += fmt.Sprintf(`
<post_title>
%s
</post_title>`, post.Post.Title)
	}

	if post.Post.Body != "" {
		message += fmt.Sprintf(`
<post_body>
%s
</post_body>`, post.Post.Body)
	}

	if opReply.Body != "" {
		message += fmt.Sprintf(`
<explanatory_comment>
%s
</explanatory_comment>`, opReply.Body)
	}

	return message
}

func suggestApprove(approval Approval) {
	fmt.Printf(`Recommendation: %s
Someone: %s
Something: %s
Consequences: %s
Explanation: %s

`,
		aurora.Green("Approve"),
		aurora.Gray(9, approval.Someone),
		aurora.Gray(9, approval.Something),
		aurora.Gray(9, approval.Consequences),
		aurora.Gray(12, approval.Explanation),
	)
}

func suggestRemove(removal Removal) {
	fmt.Printf(`Recommendation: %s
Reason: %s

`, aurora.Red("Remove"), removal.Reason)
}

func prettyPrint(i interface{}) error {
	s, err := json.MarshalIndent(i, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(s))
	return err
}
