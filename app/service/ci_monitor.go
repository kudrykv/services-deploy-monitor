package service

import (
	"context"
	"github.com/Sirupsen/logrus"
	"github.com/kudrykv/services-deploy-monitor/app/config"
	"github.com/kudrykv/services-deploy-monitor/app/internal/httputil"
	"github.com/kudrykv/services-deploy-monitor/app/internal/logging"
	"time"
)

type ciMonitor struct {
	cm config.Monitor
	ci CircleCi
	gh GhWrap
}

func NewCiMonitor(cm config.Monitor, gh GhWrap, ci CircleCi) CiMonitor {
	return &ciMonitor{
		cm: cm,
		ci: ci,
		gh: gh,
	}
}

func (s *ciMonitor) Monitor(ctx context.Context, hook AggregatedWebhook, f func(context.Context, Event)) {
	fields := logrus.Fields{
		"request_id": httputil.GetRequestId(ctx),
		"event":      hook.Event,
	}

	var event Event

	switch hook.Event {
	case PullRequestEvent:
		if *hook.PullRequestEvent.Action != "closed" || *hook.PullRequestEvent.PullRequest.Merged != true {
			logging.WithFields(fields).Info("skip pr")
			return
		}

		event = Event{
			Event:     hook.Event + "_merged",
			Org:       hook.PullRequestEvent.GetRepo().GetOwner().GetLogin(),
			Repo:      hook.PullRequestEvent.GetRepo().GetName(),
			BranchRef: hook.PullRequestEvent.GetPullRequest().GetBase().GetRef(),
			Sha:       hook.PullRequestEvent.GetPullRequest().GetMergeCommitSHA(),
			PrTitle:   hook.PullRequestEvent.GetPullRequest().GetTitle(),
			PrNumber:  hook.PullRequestEvent.GetPullRequest().GetNumber(),
		}

	case ReleaseEvent:
		event = Event{
			Event:     hook.Event,
			Org:       hook.ReleaseEvent.GetRepo().GetOwner().GetLogin(),
			Repo:      hook.ReleaseEvent.GetRepo().GetName(),
			BranchRef: hook.ReleaseEvent.GetRelease().GetTargetCommitish(),
			Tag:       hook.ReleaseEvent.GetRelease().GetTagName(),
		}

	case CreateEvent:
		if hook.CreateEvent.GetRefType() != "branch" {
			logging.WithFields(fields).Info("skip " + *hook.CreateEvent.RefType)
			return
		}

		event = Event{
			Event:     hook.Event,
			Org:       hook.CreateEvent.GetRepo().GetOwner().GetLogin(),
			Repo:      hook.CreateEvent.GetRepo().GetName(),
			BranchRef: hook.CreateEvent.GetRef(),
		}

		rc, err := s.gh.Commit(ctx, event.Org, event.Repo, event.BranchRef)
		if err != nil {
			logging.WithFields(fields).WithFields(logrus.Fields{"err": err}).Error("error fetching commit info")
			return
		}

		event.Sha = rc.GetSHA()

	default:
		logging.WithFields(fields).Error("unknown event: " + hook.Event)
		return
	}

	event.Source = sourceGithub
	f(ctx, event)

	logging.WithFields(fields).Info("start timer")
	ticker := time.NewTicker(time.Duration(s.cm.PollTimeIntervalS) * time.Second)
	defer ticker.Stop()
	skips := 0

	greens := false
	allGreensRestarts := 0
	allGreensWaitTimes := s.cm.PollForGreenBuildsTimes

	event.Source = sourceCircleCi
	for {
		<-ticker.C

		filterBranch := event.BranchRef
		if hook.Event == ReleaseEvent {
			filterBranch = ""
		}

		shaOrTag := event.Sha
		if len(shaOrTag) == 0 {
			shaOrTag = event.Tag
		}

		builds, err := s.ci.BuildsForProjectMatching(event.Org, event.Repo, filterBranch, shaOrTag)

		if err != nil {
			logging.WithFields(fields).WithFields(logrus.Fields{"err": err}).Error("fetch build from ci")
			event.BuildStatus = "fetch_failed"
			f(ctx, event)
			return
		}

		if skips > s.cm.PollForBuildsTimes {
			logging.WithFields(fields).
				WithFields(logrus.Fields{"skips": skips}).
				Error("did not find build in multiple consecutive attempts")

			event.BuildStatus = "search_failed"
			f(ctx, event)
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
				event.BuildStatus = "build_failed"
				f(ctx, event)
				return
			}
		}

		allGreen := true
		for _, build := range builds {
			isGreen := build.Status == "success" || build.Status == "fixed"
			allGreen = allGreen && isGreen
			if !isGreen {
				logging.WithFields(fields).WithFields(logrus.Fields{
					"link":   build.BuildURL,
					"status": build.Status,
				}).Info("pending build or something")
			}
		}

		if allGreensRestarts > allGreensWaitTimes {
			logging.WithFields(fields).
				WithFields(logrus.Fields{"restarts": allGreensRestarts}).
				Warn("died waiting for green result")

			event.BuildStatus = "wait_failed"
			f(ctx, event)
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
			event.BuildStatus = "success"
			f(ctx, event)
			return
		}
	}
}
