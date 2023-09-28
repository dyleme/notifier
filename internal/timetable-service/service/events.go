package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Dyleme/Notifier/internal/lib/serverrors"
	"github.com/Dyleme/Notifier/internal/timetable-service/domains"
)

type EventRepository interface {
	Add(context.Context, domains.Event) (domains.Event, error)
	List(ctx context.Context, userID int, params ListParams) ([]domains.Event, error)
	Update(ctx context.Context, event domains.Event) (domains.Event, error)
	Delete(ctx context.Context, eventID, userID int) error
	ListInPeriod(ctx context.Context, userID int, from, to time.Time, params ListParams) ([]domains.Event, error)
	Get(ctx context.Context, eventID, userID int) (domains.Event, error)
	GetNotNotified(ctx context.Context) ([]domains.Event, error)
	MarkNotified(ctx context.Context, ids []int) error
	UpdateNotificationParams(ctx context.Context, eventID, userID int, params domains.NotificationParams) (domains.NotificationParams, error)
	Delay(ctx context.Context, eventID, userID int, till time.Time) error
}

func (s *Service) CreateEvent(ctx context.Context, event domains.Event) (domains.Event, error) {
	op := "Service.CreateEvent: %w"
	var createdEvent domains.Event

	err := s.repo.Atomic(ctx, func(ctx context.Context, r Repository) error {
		var err error

		createdEvent, err = r.Events().Add(ctx, event)
		if err != nil {
			return err //nolint:wrapcheck //wrapping later
		}

		return nil
	})
	if err != nil {
		err = fmt.Errorf(op, err)
		logError(ctx, err)

		return domains.Event{}, err
	}

	return createdEvent, nil
}

func (s *Service) AddTaskToEvent(ctx context.Context, userID, taskID int, start time.Time, description string) (domains.Event, error) {
	op := "Servcie.AddTaskToEvent: %w"
	var event domains.Event

	err := s.repo.Atomic(ctx, func(ctx context.Context, r Repository) error {
		task, err := r.Tasks().Get(ctx, taskID, userID)
		if err != nil {
			return err //nolint:wrapcheck //wrapping later
		}

		event = domains.EventFromTask(task, start, description)
		event, err = r.Events().Add(ctx, event)
		if err != nil {
			return err //nolint:wrapcheck //wrapping later
		}

		return nil
	})
	if err != nil {
		err = fmt.Errorf(op, err)
		logError(ctx, err)

		return domains.Event{}, err
	}

	return event, nil
}

func (s *Service) GetEvent(ctx context.Context, userID, eventID int) (domains.Event, error) {
	op := "Service.GetEvent: %w"
	tt, err := s.repo.Events().Get(ctx, eventID, userID)
	if err != nil {
		err = fmt.Errorf(op, err)
		logError(ctx, err)

		return domains.Event{}, err
	}

	return tt, nil
}

func (s *Service) ListEvents(ctx context.Context, userID int, listParams ListParams) ([]domains.Event, error) {
	op := "Service.ListEvents: %w"
	tts, err := s.repo.Events().List(ctx, userID, listParams)
	if err != nil {
		err = fmt.Errorf(op, err)
		logError(ctx, err)

		return nil, err
	}

	return tts, nil
}

func (s *Service) ListEventsInPeriod(ctx context.Context, userID int, from, to time.Time, listParams ListParams) ([]domains.Event, error) {
	op := "Service.ListEventsInPeriod: %w"
	tts, err := s.repo.Events().ListInPeriod(ctx, userID, from, to, listParams)
	if err != nil {
		err = fmt.Errorf(op, err)
		logError(ctx, err)

		return nil, err
	}

	return tts, nil
}

type EventUpdateParams struct {
	ID          int
	UserID      int
	Text        string
	Description string
	Start       time.Time
	Done        bool
}

func (s *Service) UpdateEvent(ctx context.Context, params EventUpdateParams) (domains.Event, error) {
	op := "Service.UpdateEvent: %w"
	var res domains.Event
	err := s.repo.Atomic(ctx, func(ctx context.Context, repo Repository) error {
		e, err := repo.Events().Get(ctx, params.ID, params.UserID)
		if err != nil {
			return err //nolint:wrapcheck //wrapping later
		}

		if e.IsGettingDone(params.Done) {
			err := s.notifier.Delete(ctx, e.ID) //nolint:govet //new error
			if err != nil {
				return serverrors.NewServiceError(err)
			}
		}

		e.Text = params.Text
		e.Description = params.Description
		e.Done = params.Done
		e.Start = params.Start

		res, err = s.repo.Events().Update(ctx, e)
		if err != nil {
			return err //nolint:wrapcheck //wrapping later
		}

		return nil
	})
	if err != nil {
		err = fmt.Errorf(op, err)
		logError(ctx, err)

		return domains.Event{}, err
	}

	return res, nil
}

func (s *Service) DeleteEvent(ctx context.Context, userID, eventID int) error {
	op := "Service.DeleteEvent: %w"
	err := s.repo.Events().Delete(ctx, eventID, userID)
	if err != nil {
		err = fmt.Errorf(op, err)
		logError(ctx, err)

		return err
	}

	return nil
}

func (s *Service) DelayEvent(ctx context.Context, userID, eventID int, till time.Time) error {
	op := "Service.DelayEvent: %w"
	err := s.repo.Atomic(ctx, func(ctx context.Context, repo Repository) error {
		err := repo.Events().Delay(ctx, eventID, userID, till)
		if err != nil {
			return err //nolint:wrapcheck //wraping later
		}

		err = s.notifier.Delete(ctx, eventID)
		if err != nil {
			return err //nolint:wrapcheck //wraping later
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}
