package service

import (
	"context"
	"time"
)

type repositories struct {
	users    UserRepo
	tasks    TaskRepository
	tgImages TgImagesRepository
	events   EventsRepository
}

type Service struct {
	repos       repositories
	notifierJob NotifierJob
	tr          TxManager
}

type NotifierJob interface {
	UpdateWithTime(ctx context.Context, t time.Time)
}

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

func New(
	users UserRepo,
	tasks TaskRepository,
	tgImages TgImagesRepository,
	events EventsRepository,
	trManger TxManager,
	notifierJob NotifierJob,
) *Service {
	s := &Service{
		repos: repositories{
			users:    users,
			tasks:    tasks,
			tgImages: tgImages,
			events:   events,
		},
		notifierJob: notifierJob,
		tr:          trManger,
	}

	return s
}
