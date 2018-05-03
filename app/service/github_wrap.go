package service

import (
	"context"
	"encoding/json"
	"errors"
	github2 "github.com/google/go-github/github"
	"github.com/kudrykv/services-deploy-monitor/app/service/github"
	"golang.org/x/oauth2"
	"regexp"
	"sort"
	"strings"
)

var releaseRegexTag = regexp.MustCompile("^release-\\d+W\\d+-\\d+\\.\\d+$")
var releaseRegexBranch = regexp.MustCompile("^release-\\d+W\\d+-\\d+$")

type ghWrap struct {
	org string

	client *github2.Client
}

func NewGithub(accessToken, org string) GhWrap {
	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: accessToken,
	})

	tc := oauth2.NewClient(context.Background(), ts)
	client := github2.NewClient(tc)

	return &ghWrap{
		org:    org,
		client: client,
	}
}

func (s *ghWrap) Org() string {
	return s.org
}

type SortTag []github2.RepositoryTag

func (s SortTag) Len() int {
	return len(s)
}

func (s SortTag) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SortTag) Less(i, j int) bool {
	return strings.Compare(*s[i].Name, *s[j].Name) > 0
}

func (s *ghWrap) ListReleaseTags(ctx context.Context, repo string) ([]github2.RepositoryTag, error) {
	allTags := []*github2.RepositoryTag{}
	lo := &github2.ListOptions{}

	for {
		rt, r, err := s.client.Repositories.ListTags(ctx, s.org, repo, lo)
		if err != nil {
			return nil, err
		}

		allTags = append(allTags, rt...)

		if r.NextPage == 0 {
			break
		}

		lo.Page = r.NextPage
	}

	tags := []github2.RepositoryTag{}
	for i, tag := range allTags {
		if releaseRegexTag.MatchString(*tag.Name) {
			tags = append(tags, *allTags[i])
		}
	}

	return tags, nil
}

type SortBranch []github2.Branch

func (s SortBranch) Len() int {
	return len(s)
}

func (s SortBranch) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SortBranch) Less(i, j int) bool {
	return strings.Compare(*s[i].Name, *s[j].Name) > 0
}

func (s *ghWrap) ListReleaseBranches(ctx context.Context, repo string) ([]github2.Branch, error) {
	lo := &github2.ListOptions{}
	branches := []github2.Branch{}

	for {
		allBranches, r, err := s.client.Repositories.ListBranches(ctx, s.org, repo, lo)
		if err != nil {
			return nil, err
		}

		for idx, branch := range allBranches {
			if releaseRegexBranch.MatchString(*branch.Name) {
				branches = append(branches, *allBranches[idx])
			}
		}

		if r.NextPage == 0 {
			break
		}

		lo.Page = r.NextPage
	}

	sort.Sort(SortBranch(branches))

	return branches, nil
}

func (s *ghWrap) Compare(ctx context.Context, repo, base, head string) (*github2.CommitsComparison, error) {
	comp, _, err := s.client.Repositories.CompareCommits(ctx, s.org, repo, base, head)
	return comp, err
}

func (s *ghWrap) Commits(ctx context.Context, repo, base string, pages, perPage int) ([]*github2.RepositoryCommit, error) {
	lo := github2.ListOptions{
		PerPage: perPage,
	}
	clo := &github2.CommitsListOptions{
		SHA:         base,
		ListOptions: lo,
	}

	totalNumOfCommits := pages * perPage
	var rcAll []*github2.RepositoryCommit

	for len(rcAll) < totalNumOfCommits {
		rc, r, err := s.client.Repositories.ListCommits(ctx, s.org, repo, clo)
		if err != nil {
			return nil, err
		}

		rcAll = append(rcAll, rc...)

		if r.NextPage == 0 {
			break
		}

		clo.ListOptions.Page = r.NextPage
	}

	return rcAll, nil
}

func (s *ghWrap) Commit(ctx context.Context, org, repo, sha string) (*github2.RepositoryCommit, error) {
	if len(org) == 1 {
		org = s.org
	}

	rc, _, err := s.client.Repositories.GetCommit(ctx, org, repo, sha)

	return rc, err
}

func (s *ghWrap) IsEventSupported(event string) bool {
	switch event {
	case github.PullRequestEvent, github.ReleaseEvent, github.CreateEvent:
		return true

	default:
		return false
	}
}

func (s *ghWrap) ParseWebhook(ctx context.Context, event string, body []byte) (*github.AggregatedWebhook, error) {
	hook := github.AggregatedWebhook{
		Event: event,
	}
	var err error

	switch event {
	case github.PullRequestEvent:
		err = json.Unmarshal(body, &hook.PullRequestEvent)

	case github.ReleaseEvent:
		err = json.Unmarshal(body, &hook.ReleaseEvent)

	case github.CreateEvent:
		err = json.Unmarshal(body, &hook.CreateEvent)

	default:
		err = errors.New("unrecognized event: " + event)
	}

	return &hook, err
}
