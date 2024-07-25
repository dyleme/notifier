package service

import (
	"context"
	"time"

	trManager "github.com/avito-tech/go-transaction-manager/trm/manager"
)

type repositories struct {
	periodicTasks             PeriodicTasksRepository
	basicTasks                BasicTaskRepository
	tgImages                  TgImagesRepository
	events                    EventsRepository
	defaultNotificationParams DefaultNotificationParamsRepository
}

type Service struct {
	repos       repositories
	notifierJob NotifierJob
	tr          *trManager.Manager
}

type NotifierJob interface {
	UpdateWithTime(ctx context.Context, t time.Time)
}

func New(
	periodicTasks PeriodicTasksRepository,
	basicTasks BasicTaskRepository,
	tgImages TgImagesRepository,
	events EventsRepository,
	defaultNotificationParams DefaultNotificationParamsRepository,
	trManger *trManager.Manager,
	notifierJob NotifierJob,
) *Service {
	s := &Service{
		repos: repositories{
			periodicTasks:             periodicTasks,
			basicTasks:                basicTasks,
			tgImages:                  tgImages,
			events:                    events,
			defaultNotificationParams: defaultNotificationParams,
		},
		notifierJob: notifierJob,
		tr:          trManger,
	}

	return s
}
