package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/pkg/serverrors"
	"github.com/Dyleme/Notifier/pkg/utils/timeborders"
)

//go:generate mockgen -destination=mocks/events_mocks.go -package=mocks . EventsRepository
type EventsRepository interface {
	Add(ctx context.Context, event domains.Event) (domains.Event, error)
	List(ctx context.Context, userID int, timeBorderes timeborders.TimeBorders, listParams ListParams) ([]domains.Event, error)
	Get(ctx context.Context, id int) (domains.Event, error)
	GetLatest(ctx context.Context, taskdID int) (domains.Event, error)
	Update(ctx context.Context, event domains.Event) error
	Delete(ctx context.Context, id int) error
	ListNotSended(ctx context.Context, till time.Time) ([]domains.Event, error)
	GetNearest(ctx context.Context, till time.Time) (domains.Event, error)
	MarkSended(ctx context.Context, ids []int) error
}

func (s *Service) ListEvents(ctx context.Context, userID int, timeBorders timeborders.TimeBorders, listParams ListParams) ([]domains.Event, error) {
	var events []domains.Event
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var err error
		events, err = s.repo.Events().List(ctx, userID, timeBorders, listParams)
		if err != nil {
			return fmt.Errorf("events: list: %w", err)
		}

		return nil
	})
	if err != nil {
		logError(ctx, err)

		return nil, fmt.Errorf("tr: %w", err)
	}

	return events, nil
}

func (s *Service) GetEvent(ctx context.Context, eventID, userID int) (domains.Event, error) {
	var event domains.Event
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var err error
		event, err = s.repo.Events().Get(ctx, eventID)
		if err != nil {
			return fmt.Errorf("events: get: %w", err)
		}

		if !event.BelongsTo(userID) {
			return fmt.Errorf("event: %w", serverrors.NewBusinessLogicError("event doesn't belong to user"))
		}

		return nil
	})
	if err != nil {
		err = fmt.Errorf("tr: %w", err)
		logError(ctx, err)

		return domains.Event{}, err
	}

	return event, nil
}

func (s *Service) ChangeEventTime(ctx context.Context, eventID int, newTime time.Time, userID int) error {
	if newTime.Before(time.Now()) {
		return fmt.Errorf("time: %w", serverrors.NewBusinessLogicError("time can't be in the past"))
	}
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		ev, err := s.repo.Events().Get(ctx, eventID)
		if err != nil {
			return fmt.Errorf("events get: %w", err)
		}
		if !ev.BelongsTo(userID) {
			return fmt.Errorf("event: %w", serverrors.NewBusinessLogicError("event doesn't belong to user"))
		}

		ev.SendTime = newTime

		err = s.repo.Events().Update(ctx, ev)
		if err != nil {
			return fmt.Errorf("events update: %w", err)
		}

		s.notifierJob.UpdateWithTime(ctx, newTime)

		return nil
	})
	if err != nil {
		err = fmt.Errorf("tr: %w", err)
		logError(ctx, err)

		return err
	}

	return nil
}

func (s *Service) DeleteEvent(ctx context.Context, eventID, userID int) error {
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		ev, err := s.repo.Events().Get(ctx, eventID)
		if err != nil {
			return fmt.Errorf("events get: %w", err)
		}
		if !ev.BelongsTo(userID) {
			return fmt.Errorf("event: %w", serverrors.NewBusinessLogicError("event doesn't belong to user"))
		}

		err = s.repo.Events().Delete(ctx, eventID)
		if err != nil {
			return fmt.Errorf("events delete: %w", err)
		}

		return nil
	})
	if err != nil {
		err = fmt.Errorf("tr: %w", err)
		logError(ctx, err)

		return err
	}

	return nil
}
