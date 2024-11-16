package main

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/Sadzeih/go-reddit/reddit"
	openai "github.com/sashabaranov/go-openai"
)

type UserContext struct {
	PostTitle          string `xml:"PostTitle"`
	PostUrl            string `xml:"PostUrl"`
	PostBody           string `xml:"PostBody"`
	ExplanatoryComment string `xml:"ExplanatoryComment"`
}

func MakeUserContext(post *reddit.PostAndComments, opReply *reddit.Comment) string {
	context := UserContext{
		PostTitle:          post.Post.Title,
		PostUrl:            post.Post.URL,
		PostBody:           post.Post.Body,
		ExplanatoryComment: FormatOpReply(opReply),
	}

	contextXml, err := xml.Marshal(context)
	if err != nil {
		return fmt.Sprintf("failed to marshal user context: %v", err)
	}

	return string(contextXml)
}

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
) (*Removal, error) {
	resp, err := openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemMessage},
			{Role: openai.ChatMessageRoleAssistant, Content: automodComment.Body},
			{Role: openai.ChatMessageRoleUser, Content: MakeUserContext(post, opReply)},
		},
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
