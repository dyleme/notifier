package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/internal/domain/apperr"
	"github.com/Dyleme/Notifier/pkg/log"
	"github.com/Dyleme/Notifier/pkg/model"
)

//go:generate mockgen -destination=mocks/events_mocks.go -package=mocks . EventsRepository
type EventsRepository interface {
	Add(ctx context.Context, event domain.Sending) error
	List(ctx context.Context, userID int, params ListEventsFilterParams) ([]domain.Sending, error)
	Get(ctx context.Context, id, userID int) (domain.Sending, error)
	GetLatest(ctx context.Context, taskdID int) (domain.Sending, error)
	Update(ctx context.Context, event domain.Sending) error
	Delete(ctx context.Context, id int) error
}

type ListEventsFilterParams struct {
	TimeBorders model.TimeBorders
	ListParams  ListParams
}

func (s *Service) ListEvents(ctx context.Context, userID int, params ListEventsFilterParams) ([]domain.Sending, error) {
	var events []domain.Sending
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var err error
		events, err = s.repos.events.List(ctx, userID, params)
		if err != nil {
			return fmt.Errorf("events: list: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("tr: %w", err)
	}

	return events, nil
}

func (s *Service) GetEvent(ctx context.Context, eventID, userID int) (domain.Sending, error) {
	var event domain.Sending
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var err error
		event, err = s.repos.events.Get(ctx, eventID, userID)
		if err != nil {
			return fmt.Errorf("events: get: %w", err)
		}

		return nil
	})
	if err != nil {
		return domain.Sending{}, fmt.Errorf("tr: %w", err)
	}

	return event, nil
}

func (s *Service) ChangeEventTime(ctx context.Context, eventID int, newTime time.Time, userID int) error {
	if newTime.Before(time.Now()) {
		return fmt.Errorf("time: %w", apperr.ErrEventPastType)
	}
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		ev, err := s.repos.events.Get(ctx, eventID, userID)
		if err != nil {
			return fmt.Errorf("events get: %w", err)
		}

		ev.NextSending = newTime
		ev.OriginalSending = newTime

		err = s.repos.events.Update(ctx, ev)
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

func (s *Service) DeleteEvent(ctx context.Context, eventID, userID int) error {
	err := s.repos.events.Delete(ctx, eventID)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return fmt.Errorf("events delete[eventID=%v]: %w", eventID, apperr.NotFoundError{Object: "event"})
		}

		return fmt.Errorf("events delete[eventID=%v]: %w", eventID, err)
	}

	return nil
}

func (s *Service) ReschedulEventToTime(ctx context.Context, eventID, userID int, t time.Time) error {
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		ev, err := s.repos.events.Get(ctx, eventID, userID)
		if err != nil {
			return fmt.Errorf("events get: %w", err)
		}

		ev = ev.RescheuleToTime(t)

		err = s.repos.events.Update(ctx, ev)
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

func (s *Service) SetEventDoneStatus(ctx context.Context, eventID, userID int, done bool) error {
	log.Ctx(ctx).Debug("setting event status", "eventID", eventID, "userID", userID, "status", done)
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		event, err := s.repos.events.Get(ctx, eventID, userID)
		if err != nil {
			return fmt.Errorf("get event: %w", err)
		}

		event.Done = done

		err = s.repos.events.Update(ctx, event)
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
