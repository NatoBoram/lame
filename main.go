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

	redditCredentials := reddit.Credentials{ID: redCreds.ID, Secret: redCreds.Secret, Username: redCreds.Username, Password: redCreds.Password}
	ua := fmt.Sprintf("%s:%s:%s (by /u/NatoBoram)", runtime.GOOS, packageName, version)
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

	fmt.Printf("Logged in as %s\n", aurora.Red(user.Name))

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

func mainLoop(ctx context.Context, redditClient *reddit.Client, openaiClient *openai.Client) error {
	fmt.Print("Enter a Reddit post url: ")

	reader := bufio.NewReader(os.Stdin)
	url, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	postId, err := getPostId(url)
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

`, aurora.Bold(post.Post.Title), aurora.Gray(12, post.Post.Body), post.Post.URL)

	automodComment, err := findAutomodComment(post)
	if err != nil {
		return fmt.Errorf("failed to find AutoModerator comment: %w", err)
	}

	fmt.Printf("Found comment by %s\n", aurora.Green("u/AutoModerator"))

	_, err = redditClient.Comment.LoadMoreReplies(ctx, automodComment)
	if err != nil {
		return fmt.Errorf("failed to load more replies: %w", err)
	}

	opReply, err := findExplanatoryComment(post, automodComment)
	if err != nil {
		return err
	}

	fmt.Printf(`Found explanatory comment by %s
Body: %s

`, aurora.Red("u/"+opReply.Author), aurora.Gray(12, opReply.Body))

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
		approval, err := UnmarshalApproval([]byte(resp.Choices[0].Message.FunctionCall.Arguments))
		if err != nil {
			return fmt.Errorf("failed to unmarshal approval: %w", err)
		}
		approve(approval)

	case "remove":
		removal, err := UnmarshalRemoval([]byte(resp.Choices[0].Message.FunctionCall.Arguments))
		if err != nil {
			return fmt.Errorf("failed to unmarshal removal: %w", err)
		}
		remove(removal)
	}

	return err
}

func makeUserContext(post *reddit.PostAndComments, opReply *reddit.Comment) string {
	postBody := strings.Join(strings.Split(post.Post.Body, "\n"), "\t")
	commentBody := strings.Join(strings.Split(opReply.Body, "\n"), "\t")
	return fmt.Sprintf(`Post title: %s
Post body: %s
Explanatory comment: %s`, post.Post.Title, postBody, commentBody)
}

func approve(approval Approval) error {
	fmt.Printf(`Recommendation: %s
Explanation: %s

`, aurora.Green("Approve"), aurora.Gray(12, approval.Explanation))

	return nil
}

func remove(removal Removal) error {
	fmt.Printf(`Recommendation: %s
Reason: %s

`, aurora.Red("Remove"), removal.Reason)

	return nil
}

func prettyPrint(i interface{}) {
	s, _ := json.MarshalIndent(i, "", "\t")
	fmt.Println(string(s))
}

func lameConfigDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config dir: %w", err)
	}

	path := dir + string(os.PathSeparator) + "lame"
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return path, fmt.Errorf("failed to create `lame` config dir: %w", err)
	}

	return path, err
}
