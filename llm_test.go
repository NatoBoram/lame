package main_test

import (
	"testing"

	. "github.com/NatoBoram/lame"
	"github.com/Sadzeih/go-reddit/reddit"
)

func TestMakeUserContext(t *testing.T) {
	post := &reddit.PostAndComments{
		Post: &reddit.Post{
			Title: "Test title",
			URL:   "https://redd.it/lt8zlq",
			Body:  "Test body",
		},
		Comments: []*reddit.Comment{},
		More:     nil,
	}
	opReply := &reddit.Comment{Body: "Test comment"}
	result := MakeUserContext(post, opReply)

	expected := `<UserContext><PostTitle>Test title</PostTitle><PostUrl>https://redd.it/lt8zlq</PostUrl><PostBody>Test body</PostBody><ExplanatoryComment>Test comment</ExplanatoryComment></UserContext>`
	if expected != result {
		t.Errorf("expected %s, got %s", expected, result)
	}
}
