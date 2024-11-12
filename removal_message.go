package main

import "fmt"

var reasonToRule = map[RemovalReason]string{
	ACTUAL_ANIMAL_ATTACK:                    "* **Rule 1 :** No actual animal attacks",
	BAD_EXPLANATORY_COMMENT:                 "* **Rule 3 :** Write an explanatory comment",
	DIRECT_LINK_TO_OTHER_SUBREDDIT:          "* **Rule 7 :** No direct links to other subreddits",
	DOES_NOT_FIT_THE_SUBREDDIT:              "* **Rule 4 :** Must follow the \"Leopard ate my face\" theme",
	LEOPARD_IN_TITLE_OR_EXPLANATORY_COMMENT: "* **Rule 6 :** No \"Leopards ate my face\" in the title nor explanatory comment",
	NO_CONSEQUENCES_YET:                     "* **Rule 4 :** Must follow the \"Leopard ate my face\" theme",
	NO_EXPLANATORY_COMMENT:                  "* **Rule 3 :** Write an [explanatory comment](https://www.reddit.com/r/LeopardsAteMyFace/comments/lt8zlq)",
	UNCIVIL_BEHAVIOUR:                       "* **Rule 5 :** Be civil",
}

func formatRemovalMessage(reason RemovalReason) (string, error) {
	rule, ok := reasonToRule[reason]
	if !ok {
		return "", fmt.Errorf("no rule found for reason: %s", reason)
	}

	message := fmt.Sprintf(`Thank you for your submission! Unfortunately, it has been removed for the following reason:

%s

*If you have any questions or concerns about this removal, please feel free to [message the moderators](https://www.reddit.com/message/compose/?to=/r/LeopardsAteMyFace) thru Modmail. Thanks!*`, rule)

	return message, nil
}
