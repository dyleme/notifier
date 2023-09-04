package notifier

import (
	"context"
	"time"

	"github.com/Dyleme/Notifier/internal/lib/log"
	"github.com/Dyleme/Notifier/internal/timetable-service/models"
)

type Notifier interface {
	Notify(context.Context, models.SendingNotification) error
}

type Config struct {
	Period time.Duration
}

type Service struct {
	period        time.Duration
	notifier      Notifier
	notifications map[int]*Notification
}

func New(ctx context.Context, notifier Notifier, cfg Config) *Service {
	s := &Service{notifier: notifier, period: cfg.Period, notifications: make(map[int]*Notification)}
	go s.RunJob(ctx)
	return s
}

func (s *Service) SetNotifier(notifier Notifier) {
	s.notifier = notifier
}

type Notification struct {
	timePassed   time.Duration
	notification models.SendingNotification
}

func (s *Service) RunJob(ctx context.Context) {
	ticker := time.NewTicker(s.period)
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
	for _, n := range s.notifications {
		if n.timePassed%n.notification.Params.Period == 0 {
			err := s.notifier.Notify(ctx, n.notification)
			if err != nil {
				log.Ctx(ctx).Error("notifier error", log.Err(err))
			}
		}
		n.timePassed += s.period
	}
}

func (s *Service) Add(_ context.Context, ns []models.SendingNotification) error {
	for i := 0; i < len(ns); i++ {
		s.notifications[ns[i].TimetableTaskID] = &Notification{notification: ns[i], timePassed: 0}
	}
	return nil
}

func (s *Service) Delete(_ context.Context, timetableTaskID int) error {
	delete(s.notifications, timetableTaskID)
	return nil
}
