package service

import "github.com/google/go-github/github"

type AggregatedWebhook struct {
	Event string

	PullRequestEvent *github.PullRequestEvent
	ReleaseEvent     *github.ReleaseEvent
	CreateEvent      *github.CreateEvent
}
