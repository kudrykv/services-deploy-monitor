package notifier

import (
	"bytes"
	"context"
	"github.com/Sirupsen/logrus"
	"github.com/kudrykv/services-deploy-monitor/app/internal/logging"
)

type Notifier interface {
	Do(ctx context.Context, event map[string]string)
}

type notifier struct {
	cfg Config
}

func New(cfg Config) Notifier {
	return &notifier{
		cfg: cfg,
	}
}

func (s *notifier) Do(ctx context.Context, event map[string]string) {
	var systems *Systems
	for pattern, s := range s.cfg.Cvs.Branches {
		if pattern.MatchString(event["branch_ref"]) {
			systems = &s
			break
		}
	}

	fields := logging.WithFields(logrus.Fields{"event": event})

	if systems == nil {
		for pattern, s := range s.cfg.Cvs.Tags {
			if pattern.MatchString(event["sha_or_tag"]) {
				systems = &s
				break
			}
		}

		if systems == nil {
			fields.Info("no branch or tag match")
			return
		}
	}

	switch event["source"] {
	case "github":
		sendPack, ok := systems.Github[event["event"]]
		if !ok {
			fields.Warn("unknown event")
			return
		}

		sendPack.Message.Execute(bytes.NewBuffer(nil), event)

	default:
		fields.Error("unknown source")
	}

	fields.Info("event received")
}
