package service_test

import (
	"context"

	"github.com/Dyleme/Notifier/internal/service/service"
)

type RepositoryMock struct {
	DefaultNotificationRepo service.NotificationParamsRepository
	TasksRepo               service.BasicTaskRepository
	TgImagesRepo            service.TgImagesRepository
	PeriodicTasksRepo       service.PeriodicTasksRepository
	NotificationsRepo       service.NotificationsRepository
}

func (r *RepositoryMock) Atomic(ctx context.Context, fn func(ctx context.Context, repo service.Repository) error) error {
	return fn(ctx, r)
}

func (r *RepositoryMock) DefaultNotificationParams() service.NotificationParamsRepository {
	return r.DefaultNotificationRepo
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

func (r *RepositoryMock) Notifications() service.NotificationsRepository {
	return r.NotificationsRepo
}
