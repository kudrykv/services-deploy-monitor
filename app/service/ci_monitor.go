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
	Monitor(ctx context.Context, hook gh.AggregatedWebhook, f func(map[string]string))
}

type ciMonitor struct {
	ci CircleCi
}

func NewCiMonitor(ci CircleCi) CiMonitor {
	return &ciMonitor{
		ci: ci,
	}
}

func (s *ciMonitor) Monitor(ctx context.Context, hook gh.AggregatedWebhook, f func(map[string]string)) {
	fields := logrus.Fields{
		"request_id": httputil.GetRequestId(ctx),
		"event":      hook.Event,
	}

	var org string
	var repo string
	var branchRef string
	var shaOrTag string
	var base string

	switch hook.Event {
	case gh.PullRequestEvent:
		if *hook.PullRequestEvent.Action != "closed" || *hook.PullRequestEvent.PullRequest.Merged != true {
			logging.WithFields(fields).Info("skip pr")
			return
		}

		org = *hook.PullRequestEvent.Repo.Owner.Login
		repo = *hook.PullRequestEvent.Repo.Name
		branchRef = *hook.PullRequestEvent.PullRequest.Base.Ref
		shaOrTag = *hook.PullRequestEvent.PullRequest.MergeCommitSHA
		base = *hook.PullRequestEvent.PullRequest.Base.Ref

	case gh.ReleaseEvent:
		org = *hook.ReleaseEvent.Repo.Owner.Login
		repo = *hook.ReleaseEvent.Repo.Name
		branchRef = *hook.ReleaseEvent.Release.TargetCommitish
		shaOrTag = *hook.ReleaseEvent.Release.TagName

	case gh.CreateEvent:
		if *hook.CreateEvent.RefType != "branch" {
			logging.WithFields(fields).Info("skip " + *hook.CreateEvent.RefType)
			return
		}

		org = *hook.CreateEvent.Repo.Owner.Login
		repo = *hook.CreateEvent.Repo.Name
		branchRef = *hook.CreateEvent.Ref
		shaOrTag = branchRef
		base = *hook.CreateEvent.MasterBranch

	default:
		logging.WithFields(fields).Error("unknown event: " + hook.Event)
		return
	}

	//notification := map[string]string{
	//	"event": event,
	//	"repo":  repo,
	//	"sha":   shaOrTag,
	//	"base":  base,
	//}

	logging.WithFields(fields).Info("start timer")
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	skips := 0

	greens := false
	allGreensRestarts := 0
	allGreensWaitTimes := 20

	for {
		<-ticker.C

		filterBranch := branchRef
		if hook.Event == gh.ReleaseEvent {
			filterBranch = ""
		}

		builds, err := s.ci.BuildsForProjectMatching(org, repo, filterBranch, shaOrTag)

		if err != nil {
			logging.WithFields(fields).WithFields(logrus.Fields{"err": err}).Error("fetch build from ci")
			return
		}

		if skips > 6 {
			logging.WithFields(fields).
				WithFields(logrus.Fields{"skips": skips}).
				Error("did not find build in multiple consecutive attempts")
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
			f(map[string]string{
				"source": "ci",
				"event":  hook.Event,
				"built":  "success",
				"repo":   repo,
				"sha":    shaOrTag,
				"base":   base,
			})
			return
		}
	}
}
