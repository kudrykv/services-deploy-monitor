package notifier

import (
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
	logging.WithFields(logrus.Fields{"event": event}).Info("event received")
}
