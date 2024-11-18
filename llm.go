package main

import (
	"context"
	"fmt"

	"github.com/Sadzeih/go-reddit/reddit"
	openai "github.com/sashabaranov/go-openai"
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
) (*Removal, error) {
	resp, err := openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    model,
		Messages: MakeExplanatoryCompletion(guide, post, automodComment, opReply),
		Tools: []openai.Tool{
			{Type: openai.ToolTypeFunction, Function: &remove},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ask for another removal reason: %w", err)
	}

	_, removal, err := ToolCall(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get another removal reason: %w", err)
	}

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
