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
	notifierJob NotifierJob
	tr          *trManager.Manager
}

type NotifierJob interface {
	UpdateWithTime(ctx context.Context, t time.Time)
}

func New(repo Repository, trManger *trManager.Manager, notifierJob NotifierJob) *Service {
	s := &Service{
		repo:        repo,
		notifierJob: notifierJob,
		tr:          trManger,
	}

	return s
}
