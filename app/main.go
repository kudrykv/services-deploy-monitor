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
	"regexp"
	"text/template"
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

	pullRequestMergedTpl := template.New("pull_request_merged")
	tpl, err := pullRequestMergedTpl.Parse("*{{.Repo}}:* PR \"{{.PrTitle}} ({{.PrNumber}})\" merged to `{{.BranchRef}}`")
	if err != nil {
		panic(err)
	}

	tpl2, err := pullRequestMergedTpl.Parse("*CI: {{.Repo}}:* PR \"{{.PrTitle}} ({{.PrNumber}})\" merged to `{{.BranchRef}}`")
	if err != nil {
		panic(err)
	}

	notifierService := service.New(service.Config{
		Cvs: service.Cvs{
			Branches: map[*regexp.Regexp]service.Systems{
				regexp.MustCompile("^master$"): {
					Github: map[string]service.SendPack{
						"pull_request_merged": {
							Message: tpl,
						},
					},
					CircleCi: map[string]map[string]service.SendPack{
						"pull_request_merged": {
							"success": {
								Message: tpl2,
							},
						},
						"fetch_failed":  {},
						"search_failed": {},
						"build_failed":  {},
						"wait_failed":   {},
					},
				},
			},
		},
	})

	changelogHandler := handler.NewChangelog(changelogService)
	githubWebhookHandler := handler.NewGithubWebhook(githubService, ciMonitorService, notifierService)

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
