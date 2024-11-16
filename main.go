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
	"time"

	"github.com/Sadzeih/go-reddit/reddit"
	"github.com/logrusorgru/aurora/v4"
	openai "github.com/sashabaranov/go-openai"
	"github.com/theckman/yacspin"
)

const (
	version     = "0.0.0"
	packageName = "github.com/NatoBoram/lame"
)

func main() {
	spinner, err := yacspin.New(yacspin.Config{
		CharSet:           yacspin.CharSets[11],
		Frequency:         100 * time.Millisecond,
		Message:           "",
		StopCharacter:     "✓",
		StopColors:        []string{"fgGreen"},
		StopFailCharacter: "✗",
		StopFailColors:    []string{"fgRed"},
		StopMessage:       "Done.",
		Suffix:            " ",
	})
	if err != nil {
		panic(err)
	}

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

	spinner.Message("Logging in to Reddit...")
	spinner.Start()
	user, _, err := redditClient.Account.Info(ctx)
	if err != nil {
		spinner.StopFailMessage("Failed to log in to Reddit.")
		spinner.StopFail()
		panic(err)
	}
	spinner.StopMessage(fmt.Sprintf("Logged in as %s",
		aurora.Red("u/"+user.Name).Hyperlink("https://reddit.com/u/"+user.Name),
	))
	spinner.Stop()

	openaiCreds, err := readOpenAiCredentials(dir)
	if err != nil {
		panic(err)
	}
	config := openai.DefaultConfig(openaiCreds.Token)
	config.BaseURL = openaiCreds.BaseURL
	openaiClient := openai.NewClientWithConfig(config)

	// A guide to this sub's explanatory comment rule.
	spinner.Message("Getting the guide to this sub's explanatory comment rule...")
	spinner.Start()
	guide, _, err := redditClient.Post.Get(ctx, redCreds.Guide)
	if err != nil {
		spinner.StopFailMessage("Failed to get the explanatory comment guide.")
		spinner.StopFail()
		panic(err)
	}
	spinner.StopMessage("Got the explanatory comment guide.")
	spinner.Stop()
	fmt.Println()

	for {
		err := handleEntrypoint(ctx, redditClient, guide, openaiClient, openaiCreds.Model, spinner)
		if err != nil {
			spinner.StopFailMessage(err.Error())
			if spinner.Status() == yacspin.SpinnerStopped {
				fmt.Println(err)
			} else {
				spinner.StopFail()
			}
		}
	}
}

func handleEntrypoint(ctx context.Context,
	redditClient *reddit.Client, guide *reddit.PostAndComments,
	openaiClient *openai.Client, model string,
	spinner *yacspin.Spinner,
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

	feed := ToRedditFeed(strings.TrimSpace(input))
	if feed == "" {
		err := handlePost(ctx, redditClient, guide, openaiClient, model, input, spinner)
		if err != nil {
			return fmt.Errorf("failed to handle post url: %w", err)
		}
	} else {
		err := handleFeed(ctx, redditClient, guide, openaiClient, model, feed, spinner)
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
	spinner *yacspin.Spinner,
) error {
	var after string

	for {
		spinner.Message(fmt.Sprintf("Getting %s feed...", feed))
		spinner.Start()

		opts := MaybeOptions(after)
		posts, _, err := getFeedPosts(ctx, redditClient, feed, opts)
		if err != nil {
			return fmt.Errorf("failed to get feed posts: %w", err)
		}
		spinner.StopMessage(fmt.Sprintf("Got %s feed.", feed))
		spinner.Stop()

		err = handlePage(ctx, redditClient, guide, openaiClient, model, posts, spinner)
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
	spinner *yacspin.Spinner,
) error {
	for _, post := range posts {
		// TODO: Check for `approved`. go-reddit is missing this field and it's
		// abandoned so I can't even submit a pull request.
		if post.Stickied || post.Locked {
			continue
		}

		err := handlePost(ctx, redditClient, guide, openaiClient, model, post.ID, spinner)
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
	spinner *yacspin.Spinner,
) error {
	spinner.Message(fmt.Sprintf("Getting post %s...", postId))
	spinner.Start()
	post, _, err := redditClient.Post.Get(ctx, postId)
	if err != nil {
		return fmt.Errorf("failed to get post: %w", err)
	}
	spinner.StopMessage(fmt.Sprintf("Got post %s.", postId))
	spinner.Stop()

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
		FormatAutomoderator(automodComment),
	)

	spinner.Message(fmt.Sprintf("Loading replies to %s...", FormatAutomoderator(automodComment)))
	spinner.Start()
	_, err = redditClient.Comment.LoadMoreReplies(ctx, automodComment)
	if err != nil {
		return fmt.Errorf("failed to load more replies: %w", err)
	}
	spinner.StopMessage(fmt.Sprintf("Loaded replies to %s.", FormatAutomoderator(automodComment)))
	spinner.Stop()

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

	spinner.Message("Requesting a suggestion...")
	spinner.Start()
	resp, err := openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemMessage},
			{Role: openai.ChatMessageRoleAssistant, Content: guide.Post.Body},
			{Role: openai.ChatMessageRoleUser, Content: MakeUserContext(post, opReply)},
		},
		Tools: modTools,
	})
	if err != nil {
		return fmt.Errorf("failed to create chat completion: %w", err)
	}
	spinner.StopMessage("Got a suggestion.")
	spinner.Stop()
	fmt.Println()

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
		spinner.Message("Approving...")
		spinner.Start()
		_, err := redditClient.Moderation.Approve(ctx, post.Post.FullID)
		if err != nil {
			return fmt.Errorf("failed to approve post: %w", err)
		}
		spinner.StopMessage("Approved.")
		spinner.Stop()

	case "r", "remove":
		if removal == nil {
			spinner.Message("Getting new removal reason...")
			spinner.Start()
			removal, err = retryRemovalReason(post, automodComment, opReply, ctx, model, openaiClient)
			if err != nil {
				spinner.StopFailMessage(
					fmt.Sprintf("Failed to get another removal reason: %v\n", err),
				)
				spinner.StopFail()
			}

			if removal != nil {
				spinner.StopMessage(fmt.Sprintf("Got new removal reason: %s\n", removal.Reason))
				spinner.Stop()

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

			spinner.StopFailMessage("Did not get a new removal reason")
			spinner.StopFail()
		}

		spinner.Message("Removing...")
		spinner.Start()
		_, err := redditClient.Moderation.Remove(ctx, post.Post.FullID)
		if err != nil {
			return fmt.Errorf("failed to remove post: %w", err)
		}
		spinner.StopMessage("Removed.")
		spinner.Stop()

		if removal != nil {
			removalMessage, err := FormatRemovalMessage(removal.Reason, model)
			if err != nil {
				return fmt.Errorf("failed to format removal message: %w", err)
			}

			spinner.Message("Adding removal reason...")
			spinner.Start()
			removalComment, _, err := redditClient.Comment.Submit(ctx, post.Post.FullID, removalMessage)
			if err != nil {
				return fmt.Errorf("failed to submit removal reason: %w", err)
			}
			spinner.StopMessage(fmt.Sprintf("Removal reason added: %s\n", PermaLink(removalComment.Permalink)))
			spinner.Stop()

			spinner.Message("Distinguishing and stickying removal reason...")
			spinner.Start()
			_, err = redditClient.Moderation.DistinguishAndSticky(ctx, removalComment.FullID)
			if err != nil {
				return fmt.Errorf("failed to distinguish and sticky removal reason: %w", err)
			}
			spinner.StopMessage("Distinguished and stickied removal reason.")
		}

	case "", "s", "skip":
		fmt.Println("Skipped.")

	default:
		fmt.Println("Invalid input. Skipped.")

	}

	fmt.Println()
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
