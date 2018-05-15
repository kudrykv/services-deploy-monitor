package service

import (
	"github.com/google/go-github/github"
	"regexp"
	"text/template"
)

type AggregatedWebhook struct {
	Event string

	PullRequestEvent *github.PullRequestEvent
	ReleaseEvent     *github.ReleaseEvent
	CreateEvent      *github.CreateEvent
}

type Event struct {
	Event  string
	Source string

	Org         string
	Repo        string
	BranchRef   string
	Sha         string
	Tag         string
	RefType     string
	PrTitle     string
	PrNumber    int
	BuildStatus string
}

type Config struct {
	Cvs Cvs
}

type Cvs struct {
	Branches map[*regexp.Regexp]Systems
	Tags     map[*regexp.Regexp]Systems
}

type Systems struct {
	Github map[string]SendPack
}

type SendPack struct {
	Message *template.Template
}
