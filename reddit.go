package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/Sadzeih/go-reddit/reddit"
	"github.com/logrusorgru/aurora/v4"
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

func toRedditFeed(feed string) RedditFeed {
	switch feed {
	case "hot", "h":
		return Hot
	case "new", "n":
		return New
	case "top", "t":
		return Top
	case "rising", "r":
		return Rising
	default:
		return ""
	}
}

const (
	Hot    RedditFeed = "hot"
	New    RedditFeed = "new"
	Top    RedditFeed = "top"
	Rising RedditFeed = "rising"
)

type RedditFeed string

func getFeedPosts(
	ctx context.Context,
	redditClient *reddit.Client,

	feed RedditFeed,
	opts *reddit.ListOptions,
) ([]*reddit.Post, *reddit.Response, error) {
	switch feed {
	case Hot:
		return redditClient.Subreddit.HotPosts(ctx, "LeopardsAteMyFace", opts)
	case New:
		return redditClient.Subreddit.NewPosts(ctx, "LeopardsAteMyFace", opts)
	case Top:
		return redditClient.Subreddit.TopPosts(ctx, "LeopardsAteMyFace",
			&reddit.ListPostOptions{ListOptions: *opts},
		)
	case Rising:
		return redditClient.Subreddit.RisingPosts(ctx, "LeopardsAteMyFace", opts)
	default:
		return nil, nil, fmt.Errorf("unknown feed: %s", feed)
	}
}

func maybeOptions(after string) *reddit.ListOptions {
	if after == "" {
		return nil
	}

	return &reddit.ListOptions{
		After: after,
	}
}

func formatAutomoderator(comment *reddit.Comment) aurora.Value {
	return aurora.Green("u/" + comment.Author).Hyperlink("https://reddit.com/u/" + comment.Author)
}
