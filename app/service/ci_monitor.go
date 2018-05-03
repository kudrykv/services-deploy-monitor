package service

import (
	"context"
	"github.com/Sirupsen/logrus"
	"github.com/imdario/mergo"
	"github.com/kudrykv/services-deploy-monitor/app/config"
	"github.com/kudrykv/services-deploy-monitor/app/internal/httputil"
	"github.com/kudrykv/services-deploy-monitor/app/internal/logging"
	"github.com/kudrykv/services-deploy-monitor/app/service/github"
	"strconv"
	"time"
)

const (
	sourceGithub   = "github"
	sourceCircleCi = "circleci"
)

type ciMonitor struct {
	cm config.Monitor
	ci CircleCi
	gh Gh
}

func NewCiMonitor(cm config.Monitor, gh Gh, ci CircleCi) CiMonitor {
	return &ciMonitor{
		cm: cm,
		ci: ci,
		gh: gh,
	}
}

func (s *ciMonitor) Monitor(ctx context.Context, hook github.AggregatedWebhook, f func(context.Context, map[string]string)) {
	fields := logrus.Fields{
		"request_id": httputil.GetRequestId(ctx),
		"event":      hook.Event,
	}

	var org string
	var repo string
	var branchRef string
	var shaOrTag string
	var base string
	var refType string
	var prTitle string
	var prNumber string

	switch hook.Event {
	case github.PullRequestEvent:
		if *hook.PullRequestEvent.Action != "closed" || *hook.PullRequestEvent.PullRequest.Merged != true {
			logging.WithFields(fields).Info("skip pr")
			return
		}

		org = *hook.PullRequestEvent.Repo.Owner.Login
		repo = *hook.PullRequestEvent.Repo.Name
		branchRef = *hook.PullRequestEvent.PullRequest.Base.Ref
		shaOrTag = *hook.PullRequestEvent.PullRequest.MergeCommitSHA
		base = *hook.PullRequestEvent.PullRequest.Base.Ref
		prTitle = *hook.PullRequestEvent.PullRequest.Title
		prNumber = strconv.Itoa(*hook.PullRequestEvent.PullRequest.Number)

	case github.ReleaseEvent:
		org = *hook.ReleaseEvent.Repo.Owner.Login
		repo = *hook.ReleaseEvent.Repo.Name
		branchRef = *hook.ReleaseEvent.Release.TargetCommitish
		shaOrTag = *hook.ReleaseEvent.Release.TagName

	case github.CreateEvent:
		refType = *hook.CreateEvent.RefType
		if refType != "branch" {
			logging.WithFields(fields).Info("skip " + *hook.CreateEvent.RefType)
			return
		}

		org = *hook.CreateEvent.Repo.Owner.Login
		branchRef = *hook.CreateEvent.Ref
		repo = *hook.CreateEvent.Repo.Name
		base = *hook.CreateEvent.MasterBranch

		rc, err := s.gh.Commit(ctx, org, repo, branchRef)
		if err != nil {
			logging.WithFields(fields).WithFields(logrus.Fields{"err": err}).Error("error fetching commit info")
			return
		}

		shaOrTag = *rc.SHA

	default:
		logging.WithFields(fields).Error("unknown event: " + hook.Event)
		return
	}

	notification := map[string]string{
		"org":        org,
		"repo":       repo,
		"branch_ref": branchRef,
		"sha_or_tag": shaOrTag,
		"base":       base,
		"ref_type":   refType,
		"pr_title":   prTitle,
		"pr_number":  prNumber,
	}
	fields["notification_blueprint"] = notification

	mergeAndSend(ctx, map[string]string{
		"event":  hook.Event + "_merged",
		"source": sourceGithub,
	}, notification, f)

	logging.WithFields(fields).Info("start timer")
	ticker := time.NewTicker(time.Duration(s.cm.PollTimeIntervalS) * time.Second)
	defer ticker.Stop()
	skips := 0

	greens := false
	allGreensRestarts := 0
	allGreensWaitTimes := s.cm.PollForGreenBuildsTimes

	for {
		<-ticker.C

		filterBranch := branchRef
		if hook.Event == github.ReleaseEvent {
			filterBranch = ""
		}

		builds, err := s.ci.BuildsForProjectMatching(org, repo, filterBranch, shaOrTag)

		if err != nil {
			logging.WithFields(fields).WithFields(logrus.Fields{"err": err}).Error("fetch build from ci")
			mergeAndSend(ctx, map[string]string{
				"event":  hook.Event,
				"source": sourceCircleCi,
				"built":  "fetch_failed",
			}, notification, f)
			return
		}

		if skips > s.cm.PollForBuildsTimes {
			logging.WithFields(fields).
				WithFields(logrus.Fields{"skips": skips}).
				Error("did not find build in multiple consecutive attempts")

			mergeAndSend(ctx, map[string]string{
				"event":  hook.Event,
				"source": sourceCircleCi,
				"built":  "search_failed",
			}, notification, f)
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
				mergeAndSend(ctx, map[string]string{
					"event":  hook.Event,
					"source": sourceCircleCi,
					"built":  "build_failed",
				}, notification, f)
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

			mergeAndSend(ctx, map[string]string{
				"event":  hook.Event,
				"source": sourceCircleCi,
				"built":  "wait_failed",
			}, notification, f)
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
			mergeAndSend(ctx, map[string]string{
				"event":  hook.Event,
				"source": sourceCircleCi,
				"built":  "success",
			}, notification, f)
			return
		}
	}
}

func mergeAndSend(ctx context.Context, dest map[string]string, source map[string]string, f func(context.Context, map[string]string)) {
	fields := logrus.Fields{
		"request_id": httputil.GetRequestId(ctx),
	}

	if err := mergo.Merge(&dest, source); err != nil {
		logging.WithFields(fields).WithFields(logrus.Fields{
			"dest": dest,
			"src":  source,
		}).Error("failed to merge, dropping message")
		return
	}

	logging.WithFields(fields).WithFields(logrus.Fields{"notification": dest}).Info("passing message to notification system")
	f(ctx, dest)
}
