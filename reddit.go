package main

import (
	"fmt"
	"strings"

	"github.com/vartanbeno/go-reddit/v2/reddit"
)

// AutoModeratorID is the user ID of u/AutoModerator.
const AutoModeratorID = "t2_6l4z3"

// GetPostId gets the `lt8zlq` out of a full POST URL like this:
//
// https://www.reddit.com/r/LeopardsAteMyFace/comments/lt8zlq/a_guide_to_this_subs_explanatory_comment_rule
func GetPostId(url string) (string, error) {

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

// FindAutomodComment finds the first top-level comment made by u/AutoModerator.
func FindAutomodComment(post *reddit.PostAndComments) (*reddit.Comment, error) {
	for _, comment := range post.Comments {
		if comment.Author == "AutoModerator" {
			return comment, nil
		}
	}

	return nil, fmt.Errorf("u/AutoModerator's comment was not in the first comment page")
}

// FindExplanatoryComment finds the first reply made by the post author under u/AutoModerator's request for an explanatory comment.
func FindExplanatoryComment(post *reddit.PostAndComments, automodComment *reddit.Comment) (*reddit.Comment, error) {
	for _, comment := range automodComment.Replies.Comments {
		if comment.AuthorID == post.Post.AuthorID && comment.ParentID == fmt.Sprintf("t1_%s", automodComment.ID) {
			return comment, nil
		}
	}

	return nil, fmt.Errorf("there is no explanatory comment")
}

// PermaLink converts a Reddit "permalink" to a full URL.
func PermaLink(permalink string) string {
	return "https://reddit.com" + permalink
}
