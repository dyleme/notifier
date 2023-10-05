package service

import (
	"context"
	"time"
)

type Repository interface {
	Atomic(ctx context.Context, fn func(ctx context.Context, repo Repository) error) error
	DefaultNotificationParams() NotificationParamsRepository
	Tasks() TaskRepository
	Events() EventRepository
	TgImages() TgImagesRepository
}

type Service struct {
	repo        Repository
	notifierJob NotifierJob
	notifier    Notifier
}

type Config struct {
	CheckTasksPeriod time.Duration
}

func New(_ context.Context, repo Repository, notifier Notifier) *Service {
	s := &Service{
		repo: repo,
		notifierJob: NotifierJob{
			repo:     repo,
			notifier: notifier,
		},
		notifier: notifier,
	}

	return s
}
