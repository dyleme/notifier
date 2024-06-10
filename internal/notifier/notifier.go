package notifier

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/pkg/log"
	"github.com/Dyleme/Notifier/pkg/utils/compare"
)

//go:generate mockgen -destination=mocks/notifier_mocks.go -package=mocks . Notifier
type Notifier interface {
	Notify(ctx context.Context, notif domains.SendingNotification) error
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
	now := time.Now()
	for eventID, n := range s.notifications {
		if n.nextNotifTime.Before(now) {
			err := s.notifier.Notify(ctx, n.notification)
			if err != nil {
				log.Ctx(ctx).Error("notifier error", log.Err(err))
			}

			s.notifications[eventID].nextNotifTime = now.Add(n.notification.Params.Period)
		}
	}
	log.Ctx(ctx).Debug("notifier notify", "notifications", s.notifications)

	t := s.calcNextNotificationTime()
	s.setTimerForNextNotification(ctx, t)
}

func (s *Service) RunJob(ctx context.Context) {
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
	log.Default().Debug("notif", "notification", fmt.Sprintf("%#v", s.notifications))

	if nearestNotifTime.IsZero() {
		return time.Now().Add(s.period)
	}

	return nearestNotifTime
}

func (s *Service) setTimerForNextNotification(ctx context.Context, t time.Time) {
	log.Ctx(ctx).Debug("notifier new time", "time", t)
	s.nextNotifTime = t
	s.timer.Reset(time.Until(s.nextNotifTime))
}

func (s *Service) Add(ctx context.Context, notif domains.SendingNotification) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	notifTime := slices.MaxFunc([]time.Time{time.Now(), notif.SendTime}, compare.TimeCmpWithoutZero)
	s.notifications[notif.NotificationID] = &Notification{notification: notif, nextNotifTime: notifTime}

	log.Ctx(ctx).Debug("add", slog.Any("notification", notif))

	if notifTime.Before(s.nextNotifTime) {
		s.setTimerForNextNotification(ctx, notifTime)
	}

	return nil
}

func (s *Service) Delete(_ context.Context, notifID int) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	delete(s.notifications, notifID)

	return nil
}
