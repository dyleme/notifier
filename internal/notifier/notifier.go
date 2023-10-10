package notifier

import (
	"context"
	"sync"
	"time"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/pkg/log"
)

//go:generate mockgen -destination=mocks/notifier_mocks.go -package=mocks . Notifier
type Notifier interface {
	Notify(context.Context, domains.SendingNotification) error
}

type Config struct {
	Period time.Duration
}

type Service struct {
	notifier      Notifier
	notifications map[int]*Notification
	mx            *sync.Mutex
	nextNotifTime time.Time
	timer         *time.Timer
	period        time.Duration
}

func New(ctx context.Context, notifier Notifier, cfg Config) *Service {
	s := &Service{
		notifier:      notifier,
		period:        cfg.Period,
		nextNotifTime: time.Now().Add(cfg.Period),
		timer:         time.NewTimer(cfg.Period),
		notifications: make(map[int]*Notification),
		mx:            &sync.Mutex{},
	}
	go s.RunJob(ctx)

	return s
}

func (s *Service) SetNotifier(notifier Notifier) {
	s.notifier = notifier
}

type Notification struct {
	nextNotifTime time.Time
	notification  domains.SendingNotification
}

func (s *Service) notify(ctx context.Context) {
	s.mx.Lock()
	defer s.mx.Unlock()
	for eventID, n := range s.notifications {
		if n.nextNotifTime.Before(time.Now()) {
			err := s.notifier.Notify(ctx, n.notification)
			if err != nil {
				log.Ctx(ctx).Error("notifier error", log.Err(err))
			}

			s.notifications[eventID].nextNotifTime = n.nextNotifTime.Add(n.notification.Params.Period)
		}
	}

	t := s.calcNextNotificationTime()
	s.setTimerForNextNotification(t)
}

func (s *Service) RunJob(ctx context.Context) {
	s.notify(ctx)
	for {
		select {
		case <-s.timer.C:
			s.notify(ctx)
		case <-ctx.Done():
			s.timer.Stop()

			return
		}
	}
}

func (s *Service) calcNextNotificationTime() time.Time {
	var nearestNotifTime time.Time
	for _, n := range s.notifications {
		if nearestNotifTime.IsZero() || n.nextNotifTime.Before(nearestNotifTime) {
			nearestNotifTime = n.nextNotifTime
		}
	}

	if nearestNotifTime.IsZero() {
		return time.Now().Add(s.period)
	}

	return nearestNotifTime
}

func (s *Service) setTimerForNextNotification(t time.Time) {
	s.nextNotifTime = t
	s.timer.Reset(time.Until(s.nextNotifTime))
}

func (s *Service) Add(_ context.Context, n domains.SendingNotification) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.notifications[n.EventID] = &Notification{notification: n, nextNotifTime: n.NotificationTime}

	if n.NotificationTime.Before(s.nextNotifTime) {
		s.setTimerForNextNotification(n.NotificationTime)
	}

	return nil
}

func (s *Service) Delete(_ context.Context, eventID int) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	delete(s.notifications, eventID)

	return nil
}
