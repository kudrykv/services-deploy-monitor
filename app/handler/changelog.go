package handler

import (
	"github.com/kudrykv/services-deploy-monitor/app/service"
	"goji.io/pat"
	"net/http"
	"strconv"
)

type Changelog interface {
	Build(w http.ResponseWriter, r *http.Request)
}

type changelog struct {
	changelogService service.Changelog
}

func NewChangelog(cl service.Changelog) Changelog {
	return &changelog{
		changelogService: cl,
	}
}

func (h changelog) Build(w http.ResponseWriter, r *http.Request) {
	pagesString := r.URL.Query().Get("pages")
	pages, _ := strconv.Atoi(pagesString)

	s, err := h.changelogService.Build(r.Context(), pat.Param(r, "repo"), pages)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(s))
}
