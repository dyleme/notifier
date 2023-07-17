package service

import (
	"context"
	"log"
	"time"

	"github.com/Dyleme/Notifier/internal/notification-service/notifier/notification"
)

type Repositry interface {
	GetNewNotifications(context.Context) ([]notification.Notification, error)
	AddNotification(context.Context, notification.Notification) error
	DeleteNotification(ctx context.Context, id int) error
	GetNotification(ctx context.Context, id int) (notification.Notification, error)
	GetFutureUserNotifications(ctx context.Context, userID int) ([]notification.Notification, error)
}

type Notifier interface {
	Notify(context.Context, []notification.Notification) error
}

type Service struct {
	repo     Repositry
	notifier Notifier
}

func New(repo Repositry, notifier Notifier) *Service {
	return &Service{repo: repo, notifier: notifier}
}

func (s *Service) RunJob(ctx context.Context, period time.Duration) {
	ticker := time.NewTicker(period)
	for {
		select {
		case <-ticker.C:
			s.notify(ctx)
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (s *Service) notify(ctx context.Context) {
	ns, err := s.repo.GetNewNotifications(ctx)
	if err != nil {
		log.Printf("job: %v\n", err)
		return
	}

	err = s.notifier.Notify(ctx, ns)
	if err != nil {
		log.Printf("job: %v\n", err)
		return
	}
}

func (s *Service) AddNotification(ctx context.Context, n notification.Notification) error {
	return s.repo.AddNotification(ctx, n)
}

func (s *Service) DeleteNotification(ctx context.Context, id int) error {
	return s.repo.DeleteNotification(ctx, id)
}
