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
	Notify(ctx context.Context, notif domains.SendingEvent) error
}

type Config struct {
	Period time.Duration
}

type Service struct {
	notifier      Notifier
	events        map[int]*Event
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
		events:        make(map[int]*Event),
		mx:            &sync.Mutex{},
	}
	go s.RunJob(ctx)

	return s
}

func (s *Service) SetNotifier(notifier Notifier) {
	s.notifier = notifier
}

type Event struct {
	nextNotifTime time.Time
	event         domains.SendingEvent
}

func (s *Service) notify(ctx context.Context) {
	s.mx.Lock()
	defer s.mx.Unlock()
	now := time.Now()
	for taskID, n := range s.events {
		if n.nextNotifTime.Before(now) {
			err := s.notifier.Notify(ctx, n.event)
			if err != nil {
				log.Ctx(ctx).Error("notifier error", log.Err(err))
			}

			s.events[taskID].nextNotifTime = now.Add(n.event.Params.Period)
		}
	}
	log.Ctx(ctx).Debug("notifier notify", "events", s.events)

	t := s.calcNextEventTime()
	s.setTimerForNextEvent(ctx, t)
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

func (s *Service) calcNextEventTime() time.Time {
	var nearestNotifTime time.Time
	for _, n := range s.events {
		if nearestNotifTime.IsZero() || n.nextNotifTime.Before(nearestNotifTime) {
			nearestNotifTime = n.nextNotifTime
		}
	}
	log.Default().Debug("notif", "event", fmt.Sprintf("%#v", s.events))

	if nearestNotifTime.IsZero() {
		return time.Now().Add(s.period)
	}

	return nearestNotifTime
}

func (s *Service) setTimerForNextEvent(ctx context.Context, t time.Time) {
	log.Ctx(ctx).Debug("notifier new time", "time", t)
	s.nextNotifTime = t
	s.timer.Reset(time.Until(s.nextNotifTime))
}

func (s *Service) Add(ctx context.Context, notif domains.SendingEvent) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	notifTime := slices.MaxFunc([]time.Time{time.Now(), notif.SendTime}, compare.TimeCmpWithoutZero)
	s.events[notif.EventID] = &Event{event: notif, nextNotifTime: notifTime}

	log.Ctx(ctx).Debug("add", slog.Any("event", notif))

	if notifTime.Before(s.nextNotifTime) {
		s.setTimerForNextEvent(ctx, notifTime)
	}

	return nil
}

func (s *Service) Delete(_ context.Context, eventID int) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	delete(s.events, eventID)

	return nil
}
