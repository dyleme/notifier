package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/pkg/serverrors"
)

//go:generate mockgen -destination=mocks/events_mocks.go -package=mocks . EventRepository
type EventRepository interface { //nolint:interfacebloat //too many methods
	Add(ctx context.Context, event domains.Event) (domains.Event, error)
	List(ctx context.Context, userID int, params ListParams) ([]domains.Event, error)
	Update(ctx context.Context, event domains.Event) (domains.Event, error)
	Delete(ctx context.Context, eventID, userID int) error
	ListInPeriod(ctx context.Context, userID int, from, to time.Time, params ListParams) ([]domains.Event, error)
	Get(ctx context.Context, eventID, userID int) (domains.Event, error)
	MarkNotified(ctx context.Context, eventID int) error
	GetNearestEventSendTime(ctx context.Context) (time.Time, error)
	ListEventsBefore(ctx context.Context, sendTime time.Time) ([]domains.Event, error)
	UpdateNotificationParams(ctx context.Context, eventID, userID int, params domains.NotificationParams) (domains.NotificationParams, error)
	Delay(ctx context.Context, eventID, userID int, till time.Time) error
}

func (s *Service) CreateEvent(ctx context.Context, event domains.Event) (domains.Event, error) {
	op := "Service.CreateEvent: %w"
	var createdEvent domains.Event

	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var err error

		createdEvent, err = s.repo.Events().Add(ctx, event)
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

	s.notifierJob.UpdateWithTime(ctx, createdEvent.SendTime)

	return createdEvent, nil
}

func (s *Service) AddTaskToEvent(ctx context.Context, userID, taskID int, start time.Time, description string) (domains.Event, error) {
	op := "Servcie.AddTaskToEvent: %w"
	var event domains.Event

	err := s.tr.Do(ctx, func(ctx context.Context) error {
		task, err := s.repo.Tasks().Get(ctx, taskID, userID)
		if err != nil {
			return err //nolint:wrapcheck //wrapping later
		}

		event = domains.EventFromTask(task, start, description)
		event, err = s.repo.Events().Add(ctx, event)
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

	s.notifierJob.UpdateWithTime(ctx, event.SendTime)

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
}

func (s *Service) UpdateEvent(ctx context.Context, params EventUpdateParams) (domains.Event, error) {
	op := "Service.UpdateEvent: %w"
	var event domains.Event
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		e, err := s.repo.Events().Get(ctx, params.ID, params.UserID)
		if err != nil {
			return err //nolint:wrapcheck //wrapping later
		}

		e.Text = params.Text
		e.Description = params.Description
		e.Start = params.Start

		event, err = s.repo.Events().Update(ctx, e)
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

	s.notifierJob.UpdateWithTime(ctx, event.SendTime)

	return event, nil
}

type AbstractEvent struct {
	EventID   int
	EventType domains.EventType
	UserID    int
	Done      bool
}

func (s *Service) setEventDoneStatusBasicEvent(ctx context.Context, absEvent AbstractEvent) error {
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		e, err := s.repo.Events().Get(ctx, absEvent.EventID, absEvent.UserID)
		if err != nil {
			return fmt.Errorf("events get[eventID=%v,userID=%v]: %w", absEvent.EventID, absEvent.UserID, err)
		}

		err = s.notifier.Delete(ctx, e.ID, e.UserID)
		if err != nil {
			return fmt.Errorf("delete[eventID=%v]: %w", absEvent.EventID, err)
		}

		_, err = s.repo.Events().Update(ctx, e)
		if err != nil {
			return fmt.Errorf("events update[eventID=%v]: %w", absEvent.EventID, err)
		}

		return nil
	})
	if err != nil {
		err = fmt.Errorf("atomic: %w", err)
		logError(ctx, err)

		return err
	}

	return nil
}

func (s *Service) setEventDoneStatusPeriodicEvent(ctx context.Context, absEvent AbstractEvent) error {
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		ev, err := s.repo.PeriodicEvents().Get(ctx, absEvent.EventID, absEvent.UserID)
		if err != nil {
			return fmt.Errorf("periodic events get[eventID=%v,userID=%v]: %w", absEvent.EventID, absEvent.UserID, err)
		}

		err = s.repo.PeriodicEvents().MarkNotificationDone(ctx, ev.ID, ev.UserID)
		if err != nil {
			return fmt.Errorf("periodic events mark notified[notifID:%v]: %w", ev.Notification.ID, err)
		}

		nextNotif, err := ev.NextNotification()
		if err != nil {
			return fmt.Errorf("next notification: %w", serverrors.NewBusinessLogicError(err.Error()))
		}
		_, err = s.repo.PeriodicEvents().AddNotification(ctx, nextNotif)
		if err != nil {
			return fmt.Errorf("periodic events add notification: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("atmoic: %w", err)
	}

	return nil
}

func (s *Service) SetEventDoneStatus(ctx context.Context, absEvent AbstractEvent) error {
	switch absEvent.EventType {
	case domains.BasicEventType:
		err := s.setEventDoneStatusBasicEvent(ctx, absEvent)
		if err != nil {
			return fmt.Errorf("setEventDoneStatusBasicEvent: %w", err)
		}
	case domains.PeriodicEventType:
		err := s.setEventDoneStatusPeriodicEvent(ctx, absEvent)
		if err != nil {
			return fmt.Errorf("setEventDoneStatusPeriodicEvent: %w", err)
		}
	default:
		return fmt.Errorf("unknown eventType[%v]", absEvent.EventType)
	}

	err := s.notifier.Delete(ctx, absEvent.EventID, absEvent.UserID)
	if err != nil {
		return fmt.Errorf("delete[eventID=%v]: %w", absEvent.EventID, err)
	}

	return nil
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
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		err := s.repo.Events().Delay(ctx, eventID, userID, till)
		if err != nil {
			return err //nolint:wrapcheck //wraping later
		}

		err = s.notifier.Delete(ctx, eventID, userID)
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
