package service

import (
	"github.com/kudrykv/go-circleci"
	"net"
	"net/http"
	"time"
)

type circleCi struct {
	client *circleci.Client
}

func NewCircleCi(token string) CircleCi {
	return &circleCi{
		client: &circleci.Client{
			Token: token,
			HTTPClient: &http.Client{
				Timeout: 2 * time.Second,
				Transport: http.RoundTripper(&http.Transport{
					Proxy: http.ProxyFromEnvironment,
					DialContext: (&net.Dialer{
						Timeout:   30 * time.Second,
						KeepAlive: 30 * time.Second,
						DualStack: true,
					}).DialContext,
					MaxIdleConns:          10,
					IdleConnTimeout:       90 * time.Second,
					TLSHandshakeTimeout:   5 * time.Second,
					ExpectContinueTimeout: 1 * time.Second,
				}),
			},
		},
	}
}

func (s *circleCi) BuildsForProjectMatching(org, repo, branch, shaOrTag string) ([]circleci.Build, error) {
	builds, err := s.client.ListRecentBuildsForProject(org, repo, branch, "", 30, 0)
	if err != nil {
		return nil, err
	}

	var ret []circleci.Build
	for idx, build := range builds {
		if build.VcsRevision == shaOrTag || build.VcsTag == shaOrTag {
			ret = append(ret, *builds[idx])
		}
	}

	return ret, nil
}
