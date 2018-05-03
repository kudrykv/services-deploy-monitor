package service

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"regexp"
	"sort"
	"strings"
)

var releaseRegexTag = regexp.MustCompile("^release-\\d+W\\d+-\\d+\\.\\d+$")
var releaseRegexBranch = regexp.MustCompile("^release-\\d+W\\d+-\\d+$")

type ghWrap struct {
	org string

	client *github.Client
}

func NewGithub(accessToken, org string) GhWrap {
	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: accessToken,
	})

	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	return &ghWrap{
		org:    org,
		client: client,
	}
}

func (s *ghWrap) Org() string {
	return s.org
}

type SortTag []github.RepositoryTag

func (s SortTag) Len() int {
	return len(s)
}

func (s SortTag) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SortTag) Less(i, j int) bool {
	return strings.Compare(*s[i].Name, *s[j].Name) > 0
}

func (s *ghWrap) ListReleaseTags(ctx context.Context, repo string) ([]github.RepositoryTag, error) {
	allTags := []*github.RepositoryTag{}
	lo := &github.ListOptions{}

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

	tags := []github.RepositoryTag{}
	for i, tag := range allTags {
		if releaseRegexTag.MatchString(*tag.Name) {
			tags = append(tags, *allTags[i])
		}
	}

	return tags, nil
}

type SortBranch []github.Branch

func (s SortBranch) Len() int {
	return len(s)
}

func (s SortBranch) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SortBranch) Less(i, j int) bool {
	return strings.Compare(*s[i].Name, *s[j].Name) > 0
}

func (s *ghWrap) ListReleaseBranches(ctx context.Context, repo string) ([]github.Branch, error) {
	lo := &github.ListOptions{}
	branches := []github.Branch{}

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

func (s *ghWrap) Compare(ctx context.Context, repo, base, head string) (*github.CommitsComparison, error) {
	comp, _, err := s.client.Repositories.CompareCommits(ctx, s.org, repo, base, head)
	return comp, err
}

func (s *ghWrap) Commits(ctx context.Context, repo, base string, pages, perPage int) ([]*github.RepositoryCommit, error) {
	lo := github.ListOptions{
		PerPage: perPage,
	}
	clo := &github.CommitsListOptions{
		SHA:         base,
		ListOptions: lo,
	}

	totalNumOfCommits := pages * perPage
	var rcAll []*github.RepositoryCommit

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

func (s *ghWrap) Commit(ctx context.Context, org, repo, sha string) (*github.RepositoryCommit, error) {
	if len(org) == 1 {
		org = s.org
	}

	rc, _, err := s.client.Repositories.GetCommit(ctx, org, repo, sha)

	return rc, err
}

func (s *ghWrap) IsEventSupported(event string) bool {
	switch event {
	case PullRequestEvent, ReleaseEvent, CreateEvent:
		return true

	default:
		return false
	}
}

func (s *ghWrap) ParseWebhook(ctx context.Context, event string, body []byte) (*AggregatedWebhook, error) {
	hook := AggregatedWebhook{
		Event: event,
	}
	var err error

	switch event {
	case PullRequestEvent:
		err = json.Unmarshal(body, &hook.PullRequestEvent)

	case ReleaseEvent:
		err = json.Unmarshal(body, &hook.ReleaseEvent)

	case CreateEvent:
		err = json.Unmarshal(body, &hook.CreateEvent)

	default:
		err = errors.New("unrecognized event: " + event)
	}

	return &hook, err
}
