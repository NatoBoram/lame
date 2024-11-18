package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Sadzeih/go-reddit/reddit"
	openai "github.com/sashabaranov/go-openai"
	"github.com/theckman/yacspin"
)

func FormatOpReply(opReply *reddit.Comment) string {
	if opReply == nil {
		return ""
	}

	return opReply.Body
}

func retryRemovalReason(
	post *reddit.PostAndComments,
	automodComment *reddit.Comment,
	opReply *reddit.Comment,
	ctx context.Context,
	model string,
	openaiClient *openai.Client,
	guide *reddit.PostAndComments,
	toolCall openai.ChatCompletionResponse,
	approval *Approval,
	spinner *yacspin.Spinner,
) (*Removal, error) {
	messages := MakeExplanatoryCompletion(guide, post, automodComment, opReply)
	if approval != nil {
		messages = append(messages, toolCall.Choices[0].Message)
		messages = append(messages, openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			ToolCallID: toolCall.Choices[0].Message.ToolCalls[0].ID,
			Content: fmt.Sprintf(`1. **%s** voted for, supported or wanted to impose **%s** on other people.
2. **%s** has the consequences of **%s**.
3. As a consequence of **%s**, **%s** happened to **%s**.`,
				approval.Someone, approval.Something,
				approval.Something, approval.Consequences,
				approval.Something, approval.Consequences, approval.Someone,
			),
		})
	}

	fmt.Printf("Enter some details about this removal: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("couldn't receive additional details about the removal")
	}
	input = strings.TrimSpace(input)

	if input != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: input,
		})
	}

	spinner.Message("Getting new removal reason...")
	spinner.Start()
	resp, err := openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    model,
		Messages: messages,
		Tools: []openai.Tool{
			{Type: openai.ToolTypeFunction, Function: &remove},
		},
	})
	if err != nil {
		spinner.StopFailMessage(fmt.Sprintf("Error during chat completion request: %v", err))
		spinner.StopFail()
		return nil, nil
	}

	_, removal, err := ToolCall(resp)
	if err != nil {
		spinner.StopFailMessage(fmt.Sprintf("Couldn't parse tool call: %v", err))
		spinner.StopFail()
		return removal, nil
	}

	if removal == nil {
		spinner.StopFailMessage("Did not get a new removal reason")
		spinner.StopFail()
		return removal, nil
	}

	spinner.StopMessage(fmt.Sprintf("Got new removal reason: %s", removal.Reason))
	spinner.Stop()
	return removal, nil
}

func MakeExplanatoryCompletion(
	guide *reddit.PostAndComments,
	post *reddit.PostAndComments,
	automodComment *reddit.Comment,
	opReply *reddit.Comment,
) []openai.ChatCompletionMessage {
	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: SystemMessage},
		{Role: openai.ChatMessageRoleAssistant, Content: guide.Post.Title},
		{Role: openai.ChatMessageRoleAssistant, Content: guide.Post.Body},
	}

	if post.Post.Title != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: post.Post.Title,
		})
	}

	if post.Post.URL != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: post.Post.URL,
		})
	}

	if post.Post.Body != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: post.Post.Body,
		})
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: automodComment.Body,
	})

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: FormatOpReply(opReply),
	})

	return messages
}
