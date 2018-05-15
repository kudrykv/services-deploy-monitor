package service

import (
	"context"
	"fmt"
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
	fmt.Println(notification)
}
