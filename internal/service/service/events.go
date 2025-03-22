package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Dyleme/Notifier/internal/domain"
	serverrors "github.com/Dyleme/Notifier/internal/domain/apperr"
	"github.com/Dyleme/Notifier/pkg/log"
	"github.com/Dyleme/Notifier/pkg/model"
)

//go:generate mockgen -destination=mocks/events_mocks.go -package=mocks . EventsRepository
type EventsRepository interface {
	Add(ctx context.Context, event domain.Event) (domain.Event, error)
	List(ctx context.Context, userID int, params ListEventsFilterParams) ([]domain.Event, error)
	Get(ctx context.Context, id int) (domain.Event, error)
	GetLatest(ctx context.Context, taskdID int, taskType domain.TaskType) (domain.Event, error)
	Update(ctx context.Context, event domain.Event) error
	Delete(ctx context.Context, id int) error
}

type ListEventsFilterParams struct {
	TimeBorders model.TimeBorders
	ListParams  ListParams
	Tags        []int
}

func (s *Service) ListEvents(ctx context.Context, userID int, params ListEventsFilterParams) ([]domain.Event, error) {
	var events []domain.Event
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var err error
		events, err = s.repos.events.List(ctx, userID, params)
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

func (s *Service) GetEvent(ctx context.Context, eventID, userID int) (domain.Event, error) {
	var event domain.Event
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var err error
		event, err = s.repos.events.Get(ctx, eventID)
		if err != nil {
			return fmt.Errorf("events: get: %w", err)
		}

		if err := event.BelongsTo(userID); err != nil {
			return fmt.Errorf("belongs to: %w", serverrors.NewBusinessLogicError(err.Error()))
		}

		return nil
	})
	if err != nil {
		err = fmt.Errorf("tr: %w", err)
		logError(ctx, err)

		return domain.Event{}, err
	}

	return event, nil
}

func (s *Service) ChangeEventTime(ctx context.Context, eventID int, newTime time.Time, userID int) error {
	if newTime.Before(time.Now()) {
		return fmt.Errorf("time: %w", serverrors.NewBusinessLogicError("time can't be in the past"))
	}
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		ev, err := s.repos.events.Get(ctx, eventID)
		if err != nil {
			return fmt.Errorf("events get: %w", err)
		}

		if err := ev.BelongsTo(userID); err != nil {
			return fmt.Errorf("belongs to: %w", serverrors.NewBusinessLogicError(err.Error()))
		}

		ev.FirstSend = newTime

		err = s.repos.events.Update(ctx, ev)
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
		ev, err := s.repos.events.Get(ctx, eventID)
		if err != nil {
			return fmt.Errorf("events get[eventID=%v]: %w", eventID, err)
		}

		if err := ev.BelongsTo(userID); err != nil {
			return fmt.Errorf("belongs to: %w", serverrors.NewBusinessLogicError(err.Error()))
		}

		err = s.repos.events.Delete(ctx, eventID)
		if err != nil {
			return fmt.Errorf("events delete[eventID=%v]: %w", eventID, err)
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

func (s *Service) ReschedulEventToTime(ctx context.Context, eventID, userID int, t time.Time) error {
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		ev, err := s.repos.events.Get(ctx, eventID)
		if err != nil {
			return fmt.Errorf("events get: %w", err)
		}

		if err := ev.BelongsTo(userID); err != nil {
			return fmt.Errorf("belongs to: %w", serverrors.NewBusinessLogicError(err.Error()))
		}

		ev = ev.RescheuleToTime(t)

		err = s.repos.events.Update(ctx, ev)
		if err != nil {
			return fmt.Errorf("events update: %w", err)
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

func (s *Service) SetEventDoneStatus(ctx context.Context, eventID, userID int, done bool) error {
	log.Ctx(ctx).Debug("setting event status", "eventID", eventID, "userID", userID, "status", done)
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		event, err := s.repos.events.Get(ctx, eventID)
		if err != nil {
			return fmt.Errorf("get event: %w", err)
		}

		if err := event.BelongsTo(userID); err != nil {
			return fmt.Errorf("belongs to: %w", serverrors.NewBusinessLogicError(err.Error()))
		}

		event.Done = done

		err = s.repos.events.Update(ctx, event)
		if err != nil {
			return fmt.Errorf("update event: %w", err)
		}

		switch event.TaskType {
		case domain.BasicTaskType:
		case domain.PeriodicTaskType:
			err := s.createNewEventForPeriodicTask(ctx, event.TaskID, userID)
			if err != nil {
				return fmt.Errorf("setTaskDoneStatusPeriodicTask: %w", err)
			}
		default:
			return fmt.Errorf("unknown taskType[%v]", event.TaskType)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("tr: %w", err)
	}

	return nil
}

func (s *Service) createAndAddEvent(ctx context.Context, task domain.EventCreator, userID int) error {
	log.Ctx(ctx).Debug("adding new event", "task", task, "userID", userID)
	defParams, err := s.repos.defaultNotificationParams.Get(ctx, userID)
	if err != nil {
		return fmt.Errorf("get default params: %w", err)
	}

	event, err := domain.CreateEvent(task, defParams)
	if err != nil {
		return fmt.Errorf("create event: %w", err)
	}

	log.Ctx(ctx).Debug("add event", "event", event)
	event, err = s.repos.events.Add(ctx, event)
	if err != nil {
		return fmt.Errorf("add event: %w", err)
	}

	s.notifierJob.UpdateWithTime(ctx, event.FirstSend)

	return nil
}
