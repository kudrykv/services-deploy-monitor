package service

import (
	"context"
	"regexp"
	"strings"
)

type changelog struct {
	github Github
}

func NewChangelog(g Github) Changelog {
	return &changelog{
		github: g,
	}
}

func (s *changelog) Build(ctx context.Context, repo string, pages int) (string, error) {
	releaseBranches, err := s.github.ListReleaseBranches(ctx, repo)
	if err != nil {
		return "", err
	}

	releaseTags, err := s.github.ListReleaseTags(ctx, repo)
	if err != nil {
		return "", err
	}

	if len(releaseBranches) == 0 || len(releaseTags) == 0 {
		return "TBD: generate changelogs for repos which don't have tags or release branches yet", nil
	}

	qa2master, err := s.github.Compare(ctx, repo, *releaseBranches[0].Name, "master")
	if err != nil {
		return "", err
	}

	release2qa, err := s.github.Compare(ctx, repo, *releaseTags[0].Name, *releaseBranches[0].Name)
	if err != nil {
		return "", err
	}

	keyReleases := map[string]string{}
	for _, rt := range releaseTags {
		keyReleases[*rt.Commit.SHA] = *rt.Name
	}

	chlog := "## Dev:\n"
	for i := len(qa2master.Commits) - 1; i >= 0; i -= 1 {
		chlog += s.enrichWithPrLink(repo, shortenCommit(*qa2master.Commits[i].Commit.Message)) + "\n"
	}

	chlog += "\n## QA:"
	chlog += "\n#### " + *releaseBranches[0].Name + " (QA)\n"
	for i := len(release2qa.Commits) - 1; i >= 0; i -= 1 {
		chlog += shortenCommit(*release2qa.Commits[i].Commit.Message) + "\n"
	}

	if pages < 1 {
		pages = 1
	}
	rc, err := s.github.Commits(ctx, repo, *releaseTags[0].Name, pages, 30)
	if err != nil {
		return "", err
	}

	chlog += "\n## PROD:"
	for _, entry := range rc {
		if value, ok := keyReleases[*entry.SHA]; ok {
			chlog += "\n#### " + value + "\n"
		}

		chlog += s.enrichWithJiraLink(s.enrichWithPrLink(repo, shortenCommit(*entry.Commit.Message))) + "\n"
	}

	return chlog, nil
}

func shortenCommit(message string) string {
	return "* " + strings.Split(message, "\n")[0]
}

var prRegex = regexp.MustCompile("\\(#(\\d+)\\)")

func (s *changelog) enrichWithPrLink(repo, message string) string {
	if prs := prRegex.FindAllStringSubmatch(message, -1); len(prs) > 0 {
		for _, pr := range prs {
			message += " https://github.com/" + s.github.Org() + "/" + repo + "/pull/" + pr[1]
		}
	}

	return message
}

var jiraRegex = regexp.MustCompile("([A-Z]+-\\d+)")

func (s *changelog) enrichWithJiraLink(message string) string {
	if jiras := jiraRegex.FindAllStringSubmatch(message, -1); len(jiras) > 0 {
		for _, jira := range jiras {
			message += " https://fubotv.atlassian.net/browse/" + jira[1]
		}
	}

	return message
}
