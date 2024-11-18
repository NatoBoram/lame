package main_test

import (
	"testing"

	. "github.com/NatoBoram/lame"
	"github.com/Sadzeih/go-reddit/reddit"
	openai "github.com/sashabaranov/go-openai"
)

func TestFormatOpReply_Nil(t *testing.T) {
	result := FormatOpReply(nil)

	expected := ""
	if result != expected {
		t.Errorf("FormatOpReply(nil) = %s; expected %s", result, expected)
	}
}

func TestFormatOpReply(t *testing.T) {
	opReply := &reddit.Comment{Body: "Hello, world!"}
	result := FormatOpReply(opReply)

	expected := "Hello, world!"
	if result != expected {
		t.Errorf("FormatOpReply(%v) = %s; expected %s", opReply, result, expected)
	}
}

func TestMakeExplanatoryCompletion(t *testing.T) {
	guide := &reddit.PostAndComments{Post: &reddit.Post{
		Title: "Guide Title",
		Body:  "Guide Body",
		URL:   "Guide URL",
	}}
	post := &reddit.PostAndComments{Post: &reddit.Post{
		Title: "Post Title",
		Body:  "Post Body",
	}}
	automodComment := &reddit.Comment{Body: "Automod Comment"}
	opReply := &reddit.Comment{Body: "Op Reply"}

	result := MakeExplanatoryCompletion(guide, post, automodComment, opReply)

	expected := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: SystemMessage},
		{Role: openai.ChatMessageRoleAssistant, Content: "Guide Title"},
		{Role: openai.ChatMessageRoleAssistant, Content: "Guide Body"},
		{Role: openai.ChatMessageRoleUser, Content: "Post Title"},
		{Role: openai.ChatMessageRoleUser, Content: "Post Body"},
		{Role: openai.ChatMessageRoleAssistant, Content: "Automod Comment"},
		{Role: openai.ChatMessageRoleUser, Content: "Op Reply"},
	}

	if len(result) != len(expected) {
		t.Errorf("MakeExplanatoryCompletion(...) = %v; expected %v", result, expected)
		return
	}

	for i := range result {
		if result[i].Role != expected[i].Role || result[i].Content != expected[i].Content {
			t.Errorf("MakeExplanatoryCompletion(...)[%d] = %v; expected %v", i, result[i], expected[i])
			return
		}
	}
}
