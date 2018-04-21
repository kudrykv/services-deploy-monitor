package notifier

import (
	"github.com/kudrykv/services-deploy-monitor/app/service"
	"regexp"
	"text/template"
)

type Config struct {
	Actions    map[string]SubAction
	Slacks     map[string]SlackConfig
	SuperAnnoy SuperAnnoy
}

type SubAction struct {
	Packs map[string]Pack
}

type Pack struct {
	Message *template.Template
	Slack   service.Slack
}

type SlackConfig struct {
	Client  service.Slack
	Channel string
}

type SuperAnnoy struct {
	Annoy   map[string]RepoConfig
	Default RepoConfig
}

type RepoConfig struct {
	MatchBranch *regexp.Regexp
	MatchTag    *regexp.Regexp
}
