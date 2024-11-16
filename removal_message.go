package main

import "fmt"

var reasonToRule = map[RemovalReason]string{
	ACTUAL_ANIMAL_ATTACK:                    "* **Rule 1 :** No actual animal attacks",
	NO_EXPLANATORY_COMMENT:                  "* **Rule 3 :** Write an [explanatory comment](https://www.reddit.com/r/LeopardsAteMyFace/comments/lt8zlq)",
	DOES_NOT_FIT_THE_SUBREDDIT:              "* **Rule 4 :** Must follow the \"Leopard ate my face\" theme",
	UNCIVIL_BEHAVIOUR:                       "* **Rule 5 :** Be civil",
	LEOPARD_IN_TITLE_OR_EXPLANATORY_COMMENT: "* **Rule 6 :** No \"Leopards ate my face\" in the title nor explanatory comment",
	DIRECT_LINK_TO_OTHER_SUBREDDIT:          "* **Rule 7 :** No direct links to other subreddits",
	BAD_EXPLANATORY_COMMENT:                 "* **Rule 3 :** Write an [explanatory comment](https://www.reddit.com/r/LeopardsAteMyFace/comments/lt8zlq)\n\nYou wrote a comment, but it wasn't an [explanation](https://www.reddit.com/r/LeopardsAteMyFace/comments/lt8zlq).",
	DISTINCT_ENABLER_AND_VICTIM:             "* **Rule 4 :** Must follow the \"Leopard ate my face\" theme\n\nThe enabler and the victim must be the same person.",
	FUTURE_CONSEQUENCES:                     "* **Rule 4 :** Must follow the \"Leopard ate my face\" theme\n\nThis is not a subreddit of the future. The consequences must have already happened.",
	NO_CONSEQUENCES:                         "* **Rule 4 :** Must follow the \"Leopard ate my face\" theme\n\nThere are no consequences in your post.",
}

func FlairToRemovalReason(removalReason RemovalReason) RemovalReason {
	for _, flair := range trappedFlairs {
		if removalReason == flair {
			return DOES_NOT_FIT_THE_SUBREDDIT
		}
	}

	return removalReason
}

func FormatRemovalMessage(removalReason RemovalReason, model string) (string, error) {
	reason := FlairToRemovalReason(removalReason)
	rule, ok := reasonToRule[reason]
	if !ok {
		return "", fmt.Errorf("no rule found for reason: %s", reason)
	}

	formattedModel := fmt.Sprintf("`%s`", model)

	message := fmt.Sprintf(`Thank you for your submission! Unfortunately, it has been removed for the following reason:

%s

*This removal was LLM-assisted. See the source code at <https://github.com/NatoBoram/lame>. Model: %s.*

*If you have any questions or concerns about this removal, please feel free to [message the moderators](https://www.reddit.com/message/compose/?to=/r/LeopardsAteMyFace) thru Modmail. Thanks!*`,
		rule, formattedModel,
	)

	return message, nil
}
