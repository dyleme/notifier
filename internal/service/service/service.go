package service

import (
	"context"
	"time"

	"github.com/avito-tech/go-transaction-manager/trm"
)

type repositories struct {
	periodicTasks             PeriodicTasksRepository
	basicTasks                BasicTaskRepository
	tgImages                  TgImagesRepository
	events                    EventsRepository
	defaultNotificationParams DefaultNotificationParamsRepository
	tags                      TagsRepository
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
	Do(ctx context.Context, fn func(ctx context.Context) error) (err error)
	DoWithSettings(ctx context.Context, s trm.Settings, fn func(ctx context.Context) error) (err error)
}

func New(
	periodicTasks PeriodicTasksRepository,
	basicTasks BasicTaskRepository,
	tgImages TgImagesRepository,
	events EventsRepository,
	defaultNotificationParams DefaultNotificationParamsRepository,
	tags TagsRepository,
	trManger TxManager,
	notifierJob NotifierJob,
) *Service {
	s := &Service{
		repos: repositories{
			periodicTasks:             periodicTasks,
			basicTasks:                basicTasks,
			tgImages:                  tgImages,
			events:                    events,
			defaultNotificationParams: defaultNotificationParams,
			tags:                      tags,
		},
		notifierJob: notifierJob,
		tr:          trManger,
	}

	return s
}
