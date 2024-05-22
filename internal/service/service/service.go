package service

import (
	"context"
	"time"

	"github.com/Dyleme/Notifier/internal/domains"
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
	notifier    Notifier
	notifierJob NotifierJob
	tr          *trManager.Manager
}

type Notifier interface {
	Add(ctx context.Context, notif domains.SendingNotification) error
	Delete(ctx context.Context, eventID, userID int) error
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
