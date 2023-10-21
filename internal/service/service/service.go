package service

import (
	"context"
	"time"

	trManager "github.com/avito-tech/go-transaction-manager/trm/manager"
)

type Repository interface {
	DefaultNotificationParams() NotificationParamsRepository
	Tasks() TaskRepository
	Events() EventRepository
	TgImages() TgImagesRepository
	PeriodicEvents() PeriodicEventsRepository
}

type Service struct {
	repo        Repository
	notifierJob *NotifierJob
	notifier    Notifier
	tr          *trManager.Manager
}

type Config struct {
	CheckTasksPeriod time.Duration
}

func New(_ context.Context, repo Repository, trManger *trManager.Manager, notifier Notifier, cfg Config) *Service {
	s := &Service{
		repo:        repo,
		notifierJob: NewNotifierJob(repo, notifier, cfg, trManger),
		notifier:    notifier,
		tr:          trManger,
	}

	return s
}
