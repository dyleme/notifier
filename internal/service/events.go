package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/internal/domain/apperr"
	"github.com/Dyleme/Notifier/pkg/log"
	"github.com/Dyleme/Notifier/pkg/model"
)

//go:generate mockgen -destination=mocks/events_mocks.go -package=mocks . EventsRepository
type EventsRepository interface {
	AddSending(ctx context.Context, event domain.Sending) error
	GetSending(ctx context.Context, id int) (domain.Sending, error)
	GetLatestSending(ctx context.Context, taskdID int) (domain.Sending, error)
	GetNearest(ctx context.Context) (time.Time, error)
	UpdateSending(ctx context.Context, sending domain.Sending) error
	DeleteSending(ctx context.Context, id int) error
	Get(ctx context.Context, eventID, userID int) (domain.Event, error)
	List(ctx context.Context, userID int, params ListEventsFilterParams) ([]domain.Event, error)
	ListNotSent(ctx context.Context, till time.Time) ([]domain.Event, error)
}

type ListEventsFilterParams struct {
	TimeBorders model.TimeBorders
	ListParams  ListParams
}

func (s *Service) ListEvents(ctx context.Context, userID int, params ListEventsFilterParams) ([]domain.Event, error) {
	log.Ctx(ctx).Debug("in list events")
	var events []domain.Event
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var err error
		log.Ctx(ctx).Debug("args", slog.Any("arg", params))
		events, err = s.repos.events.List(ctx, userID, params)
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("tr: %w", err)
	}

	return events, nil
}

func (s *Service) GetEvent(ctx context.Context, eventID, userID int) (domain.Event, error) {
	var event domain.Event
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var err error
		event, err = s.repos.events.Get(ctx, eventID, userID)
		if err != nil {
			return fmt.Errorf("events: get: %w", err)
		}

		return nil
	})
	if err != nil {
		return domain.Event{}, fmt.Errorf("tr: %w", err)
	}

	return event, nil
}

func (s *Service) ChangeEventTime(ctx context.Context, eventID int, newTime time.Time) error {
	if newTime.Before(time.Now()) {
		return fmt.Errorf("time: %w", apperr.ErrEventPastType)
	}
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		ev, err := s.repos.events.GetSending(ctx, eventID)
		if err != nil {
			return fmt.Errorf("events get: %w", err)
		}

		ev.NextSending = newTime
		ev.OriginalSending = newTime

		err = s.repos.events.UpdateSending(ctx, ev)
		if err != nil {
			return fmt.Errorf("events update: %w", err)
		}

		s.notifierJob.UpdateWithTime(ctx, newTime)

		return nil
	})
	if err != nil {
		return fmt.Errorf("tr: %w", err)
	}

	return nil
}

func (s *Service) DeleteSending(ctx context.Context, sendingID int) error {
	err := s.repos.events.DeleteSending(ctx, sendingID)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return fmt.Errorf("sending delete[sendingID=%v]: %w", sendingID, apperr.NotFoundError{Object: "sending"})
		}

		return fmt.Errorf("sending delete[sendingID=%v]: %w", sendingID, err)
	}

	return nil
}

func (s *Service) ReschedulSendingToTime(ctx context.Context, sendingID int, t time.Time) error {
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		sending, err := s.repos.events.GetSending(ctx, sendingID)
		if err != nil {
			return fmt.Errorf("events get: %w", err)
		}

		sending = sending.RescheuleToTime(t)

		err = s.repos.events.UpdateSending(ctx, sending)
		if err != nil {
			return fmt.Errorf("events update: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("tr: %w", err)
	}

	return nil
}

func (s *Service) SetEventDoneStatus(ctx context.Context, sending, userID int, done bool) error {
	log.Ctx(ctx).Debug("setting event status", "eventID", sending, "userID", userID, "status", done)
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		event, err := s.repos.events.GetSending(ctx, sending)
		if err != nil {
			return fmt.Errorf("get event: %w", err)
		}

		event.Done = done

		err = s.repos.events.UpdateSending(ctx, event)
		if err != nil {
			return fmt.Errorf("update event: %w", err)
		}

		err = s.createNewEvent(ctx, event.TaskID, userID)
		if err != nil {
			return fmt.Errorf("create new event: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("tr: %w", err)
	}

	return nil
}

func (s *Service) ListNotSentEvents(ctx context.Context, till time.Time) ([]domain.Event, error) {
	return s.repos.events.ListNotSent(ctx, till)
}

func (s *Service) RescheduleSending(ctx context.Context, event domain.Event) error {
	sending := event.ExtractSending()
	log.Ctx(ctx).Debug("before reschedule", "sending", sending)
	sending = sending.Rescheule(time.Now().Round(time.Second), event.NotificationPeriod)
	log.Ctx(ctx).Debug("after reschedule", "sending", sending)

	err := s.repos.events.UpdateSending(ctx, sending)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetNearest(ctx context.Context) (time.Time, error) {
	return s.repos.events.GetNearest(ctx)
}
