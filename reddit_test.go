package main_test

import (
	"testing"

	. "github.com/NatoBoram/lame"
	"github.com/Sadzeih/go-reddit/reddit"
)

func TestGetPostId(t *testing.T) {
	url := "https://www.reddit.com/r/LeopardsAteMyFace/comments/lt8zlq/a_guide_to_this_subs_explanatory_comment_rule/"
	expected := "lt8zlq"
	actual, err := GetPostId(url)
	if err != nil {
		t.Fatalf("GetPostId(%s) returned an error: %v", url, err)
	}

	if actual != expected {
		t.Errorf("GetPostId(%s) = %s; expected %s", url, actual, expected)
	}
}

func TestPermaLink(t *testing.T) {
	permalink := "/r/LeopardsAteMyFace/comments/lt8zlq/a_guide_to_this_subs_explanatory_comment_rule/"
	expected := "https://reddit.com/r/LeopardsAteMyFace/comments/lt8zlq/a_guide_to_this_subs_explanatory_comment_rule/"
	actual := PermaLink(permalink)

	if actual != expected {
		t.Errorf("PermaLink(%s) = %s; expected %s", permalink, actual, expected)
	}
}

func TestFindAutomodComment(t *testing.T) {
	post := &reddit.PostAndComments{
		Comments: []*reddit.Comment{
			{Author: "AutoModerator"},
		},
	}
	expected := post.Comments[0]
	actual, err := FindAutomodComment(post)
	if err != nil {
		t.Fatalf("FindAutomodComment() returned an error: %v", err)
	}

	if actual != expected {
		t.Errorf("FindAutomodComment() = %v; expected %v", actual, expected)
	}
}

func TestFindExplanatoryComment(t *testing.T) {
	post := &reddit.PostAndComments{
		Post: &reddit.Post{AuthorID: "1"},
	}

	automodComment := &reddit.Comment{
		ID: "h46ywtu",
		Replies: reddit.Replies{
			Comments: []*reddit.Comment{
				{AuthorID: "1", ParentID: "t1_h46ywtu"},
			},
		},
	}

	expected := automodComment.Replies.Comments[0]
	actual, err := FindExplanatoryComment(post, automodComment)
	if err != nil {
		t.Fatalf("FindExplanatoryComment() returned an error: %v", err)
	}

	if actual != expected {
		t.Errorf("FindExplanatoryComment() = %v; expected %v", actual, expected)
	}
}

func TestToRedditFeed(t *testing.T) {
	tests := []struct {
		input    string
		expected RedditFeed
	}{
		{"hot", Hot},
		{"h", Hot},
		{"new", New},
		{"n", New},
		{"rising", Rising},
		{"r", Rising},
		{"top", Top},
		{"t", Top},
		{"unknown", ""},
	}

	for _, test := range tests {
		result := ToRedditFeed(test.input)
		if result != test.expected {
			t.Errorf("toRedditFeed(%s) = %s; expected %s", test.input, result, test.expected)
		}
	}
}
