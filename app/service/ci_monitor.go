package service

import (
	"context"
	"github.com/Sirupsen/logrus"
	"github.com/kudrykv/services-deploy-monitor/app/internal/httputil"
	"github.com/kudrykv/services-deploy-monitor/app/internal/logging"
	gh "github.com/kudrykv/services-deploy-monitor/app/service/github"
	"time"
)

type CiMonitor interface {
	Monitor(ctx context.Context, wh gh.PullRequestWebhook)
}

type ciMonitor struct {
	ci CircleCi
}

func NewCiMonitor(ci CircleCi) CiMonitor {
	return &ciMonitor{
		ci: ci,
	}
}

func (s *ciMonitor) Monitor(ctx context.Context, wh gh.PullRequestWebhook) {
	fields := logrus.Fields{
		"request_id": httputil.GetRequestId(ctx),
		"action":     wh.Action,
		"state":      wh.PullRequest.State,
		"link":       wh.PullRequest.HTMLURL,
	}

	if *wh.Action != "closed" || *wh.PullRequest.State != "closed" || *wh.PullRequest.Merged != true {
		logging.WithFields(fields).Info("skip pr")
		return
	}

	logging.WithFields(fields).Info("start timer")
	ticker := time.NewTicker(10 * time.Second)
	skips := 0

	greens := false
	allGreensRestarts := 0
	allGreensWaitTimes := 20

	for {
		<-ticker.C

		builds, err := s.ci.BuildsForProjectMatching(
			*wh.Repository.Owner.Login,
			*wh.Repository.Name,
			*wh.PullRequest.Base.Ref,
			*wh.PullRequest.Base.SHA,
		)

		if err != nil {
			logging.WithFields(fields).WithFields(logrus.Fields{"err": err}).Error("fetch build from ci")
			ticker.Stop()
			return
		}

		if skips > 6 {
			logrus.WithFields(fields).
				WithFields(logrus.Fields{"skips": skips}).
				Error("did not find build in multiple consecutive attempts")
			ticker.Stop()
			return
		}

		if len(builds) == 0 {
			logging.WithFields(fields).WithFields(logrus.Fields{"skips": skips}).Warn("did not find build")
			skips += 1
			continue
		}

		for _, build := range builds {
			if build.Status == "canceled" || build.Status == "failed" {
				logging.WithFields(fields).Info("build failed in circleci")
				ticker.Stop()
				return
			}
		}

		allGreen := true
		for _, build := range builds {
			isGreen := build.Status == "success" || build.Status == "fixed"
			allGreen = allGreen && isGreen
			if !isGreen {
				logrus.WithFields(fields).WithFields(logrus.Fields{
					"link":   build.BuildURL,
					"status": build.Status,
				}).Info("pending build or something")
			}
		}

		if allGreensRestarts > allGreensWaitTimes {
			logging.WithFields(fields).
				WithFields(logrus.Fields{"restarts": allGreensRestarts}).
				Warn("died waiting for green result")
			ticker.Stop()
			return
		}

		if !allGreen {
			logging.WithFields(fields).
				WithFields(logrus.Fields{"restarts": allGreensRestarts}).
				Info("some builds not green, restart")
			allGreensRestarts += 1
			greens = false
			continue
		}

		if !greens {
			greens = allGreen
			logging.WithFields(fields).Info("all greens. restart once to make sure")
		} else {
			// successfully checked that all builds are green
			logging.WithFields(fields).Info("build is green")
			ticker.Stop()
			return
		}
	}
}
