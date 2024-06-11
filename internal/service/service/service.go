package service

import (
	"context"
	"time"

	trManager "github.com/avito-tech/go-transaction-manager/trm/manager"
)

type Repository interface {
	DefaultEventParams() NotificationParamsRepository
	Tasks() BasicTaskRepository
	TgImages() TgImagesRepository
	PeriodicTasks() PeriodicTasksRepository
	Events() EventsRepository
}

type Service struct {
	repo        Repository
	notifier    Notifier
	notifierJob NotifierJob
	tr          *trManager.Manager
}

type Notifier interface {
	Delete(ctx context.Context, eventID int) error
}

type NotifierJob interface {
	UpdateWithTime(ctx context.Context, t time.Time)
}

func New(repo Repository, trManger *trManager.Manager, notifier Notifier, notifierJob NotifierJob) *Service {
	s := &Service{
		repo:        repo,
		notifierJob: notifierJob,
		notifier:    notifier,
		tr:          trManger,
	}

	return s
}
