package github

import "github.com/google/go-github/github"

type PullRequestWebhook struct {
	Action      *string             `json:"action"`
	Number      *int                `json:"number"`
	PullRequest *github.PullRequest `json:"pull_request"`
	Repository  *github.Repository  `json:"repository"`
}
