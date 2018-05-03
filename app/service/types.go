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
	Action    string
	SubAction string

	Repo Repo
	Pr   *Pr
}

type Repo struct {
	Org    string
	Name   string
	Branch string
	Tag    string
	Sha    string
}

type Pr struct {
	Base   string
	Title  string
	Number int
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
