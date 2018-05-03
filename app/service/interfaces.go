package service

import (
	"context"
	github2 "github.com/google/go-github/github"
	"github.com/kudrykv/go-circleci"
	"github.com/kudrykv/services-deploy-monitor/app/service/github"
)

type Changelog interface {
	Build(ctx context.Context, repo string, pages int) (string, error)
}

type CiMonitor interface {
	Monitor(ctx context.Context, hook github.AggregatedWebhook, f func(context.Context, map[string]string))
}

type CircleCi interface {
	BuildsForProjectMatching(org, repo, branch, sha string) ([]circleci.Build, error)
}

type GhWrap interface {
	Org() string
	ListReleaseTags(ctx context.Context, repo string) ([]github2.RepositoryTag, error)
	ListReleaseBranches(ctx context.Context, repo string) ([]github2.Branch, error)
	Compare(ctx context.Context, repo, base, head string) (*github2.CommitsComparison, error)
	Commits(ctx context.Context, repo, base string, pages, perPage int) ([]*github2.RepositoryCommit, error)
	Commit(ctx context.Context, org, repo, sha string) (*github2.RepositoryCommit, error)
	IsEventSupported(event string) bool
	ParseWebhook(ctx context.Context, event string, body []byte) (*github.AggregatedWebhook, error)
}

type Slack interface {
}
