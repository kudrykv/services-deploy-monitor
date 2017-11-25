package main

import (
	"github.com/caarlos0/env"
	"github.com/kudrykv/services-deploy-monitor/app/config"
	"github.com/kudrykv/services-deploy-monitor/app/handler"
	"github.com/kudrykv/services-deploy-monitor/app/internal/httputil"
	"github.com/kudrykv/services-deploy-monitor/app/service"
	"goji.io"
	"goji.io/pat"
	"net/http"
)

func main() {
	cfg := config.Config{}
	env.Parse(&cfg.Server)
	env.Parse(&cfg.Github)
	env.Parse(&cfg.CircleCi)
	env.Parse(&cfg.Monitor)

	githubService := service.NewGithub(cfg.Github.Key, cfg.Github.Org)
	changelogService := service.NewChangelog(githubService)
	circleCiService := service.NewCircleCi(cfg.CircleCi.Key)
	ciMonitorService := service.NewCiMonitor(cfg.Monitor, githubService, circleCiService)

	changelogHandler := handler.NewChangelog(changelogService)
	githubWebhookHandler := handler.NewGithubWebhook(githubService, ciMonitorService)

	mux := goji.NewMux()
	mux.Use(trackDecorator)

	mux.HandleFunc(pat.Get("/changelog/:repo"), changelogHandler.Build)
	mux.HandleFunc(pat.Post("/webhook/github"), githubWebhookHandler.HandlePullRequest)

	http.ListenAndServe(":"+cfg.Server.Port, mux)
}

func trackDecorator(hf http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(httputil.AddRequestId(r.Context(), r))

		hf.ServeHTTP(w, r)
	})
}
