package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/pkg/log"
	"github.com/Dyleme/Notifier/pkg/serverrors"
	"github.com/Dyleme/Notifier/pkg/utils/timeborders"
)

//go:generate mockgen -destination=mocks/events_mocks.go -package=mocks . EventsRepository
type EventsRepository interface {
	Add(ctx context.Context, event domains.Event) (domains.Event, error)
	List(ctx context.Context, userID int, timeBorderes timeborders.TimeBorders, listParams ListParams) ([]domains.Event, error)
	Get(ctx context.Context, id int) (domains.Event, error)
	GetLatest(ctx context.Context, taskdID int, taskType domains.TaskType) (domains.Event, error)
	Update(ctx context.Context, event domains.Event) error
	Delete(ctx context.Context, id int) error
	ListNotSended(ctx context.Context, till time.Time) ([]domains.Event, error)
	GetNearest(ctx context.Context) (domains.Event, error)
	BatchUpdate(ctx context.Context, events []domains.Event) error
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

		if err := event.BelongsTo(userID); err != nil {
			return fmt.Errorf("belongs to: %w", serverrors.NewBusinessLogicError(err.Error()))
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

		if err := ev.BelongsTo(userID); err != nil {
			return fmt.Errorf("belongs to: %w", serverrors.NewBusinessLogicError(err.Error()))
		}

		ev.NextSendTime = newTime

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

		if err := ev.BelongsTo(userID); err != nil {
			return fmt.Errorf("belongs to: %w", serverrors.NewBusinessLogicError(err.Error()))
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

func (s *Service) ReschedulEventToTime(ctx context.Context, eventID, userID int, t time.Time) error {
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		ev, err := s.repo.Events().Get(ctx, eventID)
		if err != nil {
			return fmt.Errorf("events get: %w", err)
		}

		if err := ev.BelongsTo(userID); err != nil {
			return fmt.Errorf("belongs to: %w", serverrors.NewBusinessLogicError(err.Error()))
		}

		ev = ev.RescheuleToTime(t)

		err = s.repo.Events().Update(ctx, ev)
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
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		event, err := s.repo.Events().Get(ctx, eventID)
		if err != nil {
			return fmt.Errorf("get event: %w", err)
		}

		if err := event.BelongsTo(userID); err != nil {
			return fmt.Errorf("belongs to: %w", serverrors.NewBusinessLogicError(err.Error()))
		}

		event.Done = done

		err = s.repo.Events().Update(ctx, event)
		if err != nil {
			return fmt.Errorf("update event: %w", err)
		}

		switch event.TaskType {
		case domains.BasicTaskType:
		case domains.PeriodicTaskType:
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

func (s *Service) createAndAddEvent(ctx context.Context, task domains.EventCreator, userID int) error {
	defParams, err := s.repo.DefaultEventParams().Get(ctx, userID)
	if err != nil {
		return fmt.Errorf("get default params: %w", err)
	}

	event, err := domains.CreateEvent(task, defParams)
	if err != nil {
		return fmt.Errorf("create event: %w", err)
	}

	log.Ctx(ctx).Debug("add event", "event", event)
	event, err = s.repo.Events().Add(ctx, event)
	if err != nil {
		return fmt.Errorf("add event: %w", err)
	}

	s.notifierJob.UpdateWithTime(ctx, event.NextSendTime)

	return nil
}
