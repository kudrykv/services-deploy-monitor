package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/kudrykv/services-deploy-monitor/app/internal/logging"
)

type notifier struct {
	cfg Config
}

func New(cfg Config) Notifier {
	return &notifier{
		cfg: cfg,
	}
}

func (s *notifier) Do(ctx context.Context, notification Event) {
	switch notification.Event {
	case "pull_request_merged":
		switch notification.Source {
		case sourceGithub:
			systems := findSystems(s.cfg.Cvs, notification.BranchRef, notification.Tag)

			if systems == nil {
				logging.WithFields(logrus.Fields{"notification": notification}).Info("skip systems")
				return
			}

			sendPack, ok := systems.Github[notification.Event]
			if !ok {
				logging.WithFields(logrus.Fields{"notification": notification}).Warn("no action defined for event")
				return
			}

			fmt.Println("inside", sendPack.Message)
			buff := bytes.NewBuffer(nil)
			if err := sendPack.Message.Execute(buff, notification); err != nil {
				logging.WithFields(logrus.Fields{"notification": notification, "err": err}).Error("execute")
				return
			}

			fmt.Println("source-github-send", buff.String())

		case sourceCircleCi:
			systems := findSystems(s.cfg.Cvs, notification.BranchRef, notification.Tag)

			if systems == nil {
				logging.WithFields(logrus.Fields{"notification": notification}).Info("skip systems")
				return
			}

			eventer, ok := systems.CircleCi[notification.Event]
			if !ok {
				logging.WithFields(logrus.Fields{"notification": notification}).Warn("no action defined for event")
				return
			}

			sendPack, ok := eventer[notification.BuildStatus]
			if !ok {
				logging.WithFields(logrus.Fields{"notification": notification}).Warn("no action defined for event")
				return
			}

			buff := bytes.NewBuffer(nil)
			if err := sendPack.Message.Execute(buff, notification); err != nil {
				logging.WithFields(logrus.Fields{"notification": notification, "err": err}).Error("execute")
				return
			}

			fmt.Println("source-ci-send", buff.String())

		default:
			logging.WithFields(logrus.Fields{"notification": notification}).Error("unknown source")
		}

	default:
		logging.WithFields(logrus.Fields{"notification": notification}).Error("unknown event")
	}

	fmt.Println("notification", notification)
}

func findSystems(cvs Cvs, branch, tag string) *Systems {
	for rxp, s := range cvs.Branches {
		if rxp.MatchString(branch) {
			return &s
		}
	}

	for rxp, s := range cvs.Tags {
		if rxp.MatchString(tag) {
			return &s
		}
	}

	return nil
}
