package service

import (
	"context"
	"github.com/google/go-github/github"
	"github.com/kudrykv/go-circleci"
)

type Changelog interface {
	Build(ctx context.Context, repo string, pages int) (string, error)
}

type CiMonitor interface {
	Monitor(ctx context.Context, hook AggregatedWebhook, f func(context.Context, Event))
}

type CircleCi interface {
	BuildsForProjectMatching(org, repo, branch, sha string) ([]circleci.Build, error)
}

type GhWrap interface {
	Org() string
	ListReleaseTags(ctx context.Context, repo string) ([]github.RepositoryTag, error)
	ListReleaseBranches(ctx context.Context, repo string) ([]github.Branch, error)
	Compare(ctx context.Context, repo, base, head string) (*github.CommitsComparison, error)
	Commits(ctx context.Context, repo, base string, pages, perPage int) ([]*github.RepositoryCommit, error)
	Commit(ctx context.Context, org, repo, sha string) (*github.RepositoryCommit, error)
	IsEventSupported(event string) bool
	ParseWebhook(ctx context.Context, event string, body []byte) (*AggregatedWebhook, error)
}

type Notifier interface {
	Do(context.Context, Event)
}

type Slack interface {
}
