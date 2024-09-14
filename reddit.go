package main

import (
	"fmt"
	"strings"

	"github.com/vartanbeno/go-reddit/v2/reddit"
)

func getPostId(url string) (string, error) {
	// Get the `lt8zlq` out of a full POST URL like this:
	// https://www.reddit.com/r/LeopardsAteMyFace/comments/lt8zlq/a_guide_to_this_subs_explanatory_comment_rule

	segments := strings.Split(url, "/")
	index := 0
	for i, segment := range segments {
		if segment == "comments" {
			index = i
			break
		}
	}

	if index > 0 && index < len(segments) {
		return segments[index+1], nil
	}
	return "", fmt.Errorf("failed to get post id from url: %s", url)
}

func findAutomodComment(post *reddit.PostAndComments) (*reddit.Comment, error) {
	for _, comment := range post.Comments {
		if comment.Author == "AutoModerator" {
			return comment, nil
		}
	}

	return nil, fmt.Errorf("failed to find u/AutoModerator comment")
}

func findExplanatoryComment(post *reddit.PostAndComments, automodComment *reddit.Comment) (*reddit.Comment, error) {
	for _, comment := range automodComment.Replies.Comments {
		if comment.AuthorID == post.Post.AuthorID && comment.ParentID == fmt.Sprintf("t1_%s", automodComment.ID) {
			return comment, nil
		}
	}

	return nil, fmt.Errorf("failed to find the explanatory comment")
}
