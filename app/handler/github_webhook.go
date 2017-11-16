package handler

import (
	"context"
	"github.com/kudrykv/services-deploy-monitor/app/internal/httputil"
	"github.com/kudrykv/services-deploy-monitor/app/service"
	"net/http"
)

type GithubWebhook interface {
	HandlePullRequest(w http.ResponseWriter, r *http.Request)
}

type githubWebhook struct {
	gs service.Github
	cm service.CiMonitor
}

func NewGithubWebhook(gs service.Github, cm service.CiMonitor) GithubWebhook {
	return &githubWebhook{
		gs: gs,
		cm: cm,
	}
}

func (h githubWebhook) HandlePullRequest(w http.ResponseWriter, r *http.Request) {
	event := r.Header.Get("X-GitHub-Event")
	if event != "pull_request" && event != "release" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("unsupported event"))
		return
	}

	bytes, err := httputil.ReadBytes(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	hook, err := h.gs.ParseWebhook(r.Context(), event, bytes)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(err.Error()))
		return
	}

	httputil.Json(r.Context(), w, http.StatusOK, "OK")

	ctx := httputil.AddCustomRequestId(context.Background(), httputil.GetRequestId(r.Context()))
	go h.cm.Monitor(ctx, *hook, func(strings map[string]string) {

	})
}
