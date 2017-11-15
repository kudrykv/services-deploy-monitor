package main

import (
	"github.com/caarlos0/env"
	"github.com/kudrykv/services-deploy-monitor/app/handler"
	"github.com/kudrykv/services-deploy-monitor/app/service"
	"goji.io"
	"goji.io/pat"
	"net/http"
)

func main() {
	cfg := Config{}
	env.Parse(&cfg.Server)
	env.Parse(&cfg.Github)
	env.Parse(&cfg.CircleCi)

	githubService := service.NewGithub(cfg.Github.Key, cfg.Github.Org)
	changelogService := service.NewChangelog(githubService)
	circleCiService := service.NewCircleCi(cfg.CircleCi.Key, cfg.CircleCi.Org)
	ciMonitorService := service.NewCiMonitor(circleCiService)

	changelogHandler := handler.NewChangelog(changelogService)
	githubWebhookHandler := handler.NewGithubWebhook(githubService, ciMonitorService)

	mux := goji.NewMux()

	mux.HandleFunc(pat.Get("/changelog/:repo"), changelogHandler.Build)
	mux.HandleFunc(pat.Post("/webhook/github"), githubWebhookHandler.HandlePullRequest)

	http.ListenAndServe(":"+cfg.Server.Port, mux)
}
