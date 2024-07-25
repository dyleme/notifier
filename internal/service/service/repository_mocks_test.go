package service_test

import (
	"github.com/Dyleme/Notifier/internal/service/service"
)

type RepositoryMock struct {
	DefaultEventRepo  service.DefaultNotificationParamsRepository
	TasksRepo         service.BasicTaskRepository
	TgImagesRepo      service.TgImagesRepository
	PeriodicTasksRepo service.PeriodicTasksRepository
	EventsRepo        service.EventsRepository
}

func (r *RepositoryMock) DefaultEventParams() service.DefaultNotificationParamsRepository {
	return r.DefaultEventRepo
}

func (r *RepositoryMock) Tasks() service.BasicTaskRepository {
	return r.TasksRepo
}

func (r *RepositoryMock) TgImages() service.TgImagesRepository {
	return r.TgImagesRepo
}

func (r *RepositoryMock) PeriodicTasks() service.PeriodicTasksRepository {
	return r.PeriodicTasksRepo
}

func (r *RepositoryMock) Events() service.EventsRepository {
	return r.EventsRepo
}
