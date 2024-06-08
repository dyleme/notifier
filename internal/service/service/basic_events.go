package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/pkg/serverrors"
)

//go:generate mockgen -destination=mocks/events_mocks.go -package=mocks . EventRepository
type BasicEventRepository interface {
	Add(ctx context.Context, event domains.BasicEvent) (domains.BasicEvent, error)
	List(ctx context.Context, userID int, params ListParams) ([]domains.BasicEvent, error)
	Update(ctx context.Context, event domains.BasicEvent) (domains.BasicEvent, error)
	Delete(ctx context.Context, eventID int) error
	Get(ctx context.Context, eventID int) (domains.BasicEvent, error)
}

func (s *Service) CreateEvent(ctx context.Context, event domains.BasicEvent) (domains.BasicEvent, error) {
	op := "Service.CreateEvent: %w"
	var createdEvent domains.BasicEvent

	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var err error

		createdEvent, err = s.repo.Events().Add(ctx, event)
		if err != nil {
			return fmt.Errorf("add event: %w", err)
		}

		notif := createdEvent.NewNotification()
		_, err = s.repo.Notifications().Add(ctx, notif)
		if err != nil {
			return fmt.Errorf("add notification: %w", err)
		}

		return nil
	})
	if err != nil {
		err = fmt.Errorf(op, err)
		logError(ctx, err)

		return domains.BasicEvent{}, err
	}

	s.notifierJob.UpdateWithTime(ctx, createdEvent.Start)

	return createdEvent, nil
}

func (s *Service) AddTaskToEvent(ctx context.Context, userID, taskID int, start time.Time, description string) (domains.BasicEvent, error) {
	op := "Servcie.AddTaskToEvent: %w"
	var event domains.BasicEvent

	err := s.tr.Do(ctx, func(ctx context.Context) error {
		task, err := s.repo.Tasks().Get(ctx, taskID, userID)
		if err != nil {
			return err //nolint:wrapcheck //wrapping later
		}

		event = domains.BasicEventFromTask(task, start, description)
		event, err = s.repo.Events().Add(ctx, event)
		if err != nil {
			return fmt.Errorf("add event: %w", err)
		}

		notif := event.NewNotification()
		_, err = s.repo.Notifications().Add(ctx, notif)
		if err != nil {
			return fmt.Errorf("add notification: %w", err)
		}

		return nil
	})
	if err != nil {
		err = fmt.Errorf(op, err)
		logError(ctx, err)

		return domains.BasicEvent{}, err
	}

	s.notifierJob.UpdateWithTime(ctx, event.Start)

	return event, nil
}

func (s *Service) GetEvent(ctx context.Context, userID, eventID int) (domains.BasicEvent, error) {
	op := "Service.GetEvent: %w"
	tt, err := s.repo.Events().Get(ctx, eventID)
	if err != nil {
		err = fmt.Errorf(op, err)
		logError(ctx, err)

		return domains.BasicEvent{}, err
	}

	if !tt.BelongsTo(userID) {
		return domains.BasicEvent{}, fmt.Errorf("belongs to: %w", serverrors.NewNotFoundError(err, "event"))
	}

	return tt, nil
}

func (s *Service) ListEvents(ctx context.Context, userID int, listParams ListParams) ([]domains.BasicEvent, error) {
	op := "Service.ListEvents: %w"
	tts, err := s.repo.Events().List(ctx, userID, listParams)
	if err != nil {
		err = fmt.Errorf(op, err)
		logError(ctx, err)

		return nil, err
	}

	return tts, nil
}

func (s *Service) UpdateBasicEvent(ctx context.Context, params domains.BasicEvent, userID int) (domains.BasicEvent, error) {
	var event domains.BasicEvent
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		e, err := s.repo.Events().Get(ctx, params.ID)
		if err != nil {
			return fmt.Errorf("get event: %w", err)
		}

		if !e.BelongsTo(userID) {
			return serverrors.NewNotFoundError(err, "event")
		}

		e.Text = params.Text
		e.Description = params.Description
		e.Start = params.Start

		event, err = s.repo.Events().Update(ctx, e)
		if err != nil {
			return fmt.Errorf("update event: %w", err)
		}

		notif, err := s.repo.Notifications().GetLatest(ctx, e.ID)
		if err != nil {
			return fmt.Errorf("get latest notification: %w", err)
		}

		notif.Text = params.Text
		notif.Description = params.Description
		notif.SendTime = params.Start

		err = s.repo.Notifications().Update(ctx, notif)
		if err != nil {
			return fmt.Errorf("update notification: %w", err)
		}

		return nil
	})
	if err != nil {
		err = fmt.Errorf("tr: %w", err)
		logError(ctx, err)

		return domains.BasicEvent{}, err
	}

	s.notifierJob.UpdateWithTime(ctx, event.Start)

	return event, nil
}

func (s *Service) createNewNotificationForPeriodicEvent(ctx context.Context, eventID, userID int) error {
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		ev, err := s.repo.PeriodicEvents().Get(ctx, eventID)
		if err != nil {
			return fmt.Errorf("periodic events get[eventID=%v,userID=%v]: %w", eventID, userID, err)
		}
		if !ev.BelongsTo(userID) {
			return serverrors.NewBusinessLogicError("event does not belong to user")
		}

		nextNotif, err := ev.NewNotification(time.Now())
		if err != nil {
			return fmt.Errorf("new notification: %w", serverrors.NewBusinessLogicError(err.Error()))
		}

		_, err = s.repo.Notifications().Add(ctx, nextNotif)
		if err != nil {
			return fmt.Errorf("periodic events add notification: %w", err)
		}

		s.notifierJob.UpdateWithTime(ctx, nextNotif.SendTime)

		return nil
	})
	if err != nil {
		return fmt.Errorf("tr: %w", err)
	}

	return nil
}

func (s *Service) SetNotificationDoneStatus(ctx context.Context, notifID, userID int, done bool) error {
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		notif, err := s.repo.Notifications().Get(ctx, notifID)
		if err != nil {
			return fmt.Errorf("get notification: %w", err)
		}
		notif.Done = done
		err = s.repo.Notifications().Update(ctx, notif)
		if err != nil {
			return fmt.Errorf("update notification: %w", err)
		}
		err = s.notifier.Delete(ctx, notif.ID)
		if err != nil {
			return fmt.Errorf("delete notifier notification: %w", err)
		}
		switch notif.EventType {
		case domains.BasicEventType:
		case domains.PeriodicEventType:
			err := s.createNewNotificationForPeriodicEvent(ctx, notif.EventID, userID)
			if err != nil {
				return fmt.Errorf("setEventDoneStatusPeriodicEvent: %w", err)
			}
		default:
			return fmt.Errorf("unknown eventType[%v]", notif.EventType)
		}

		err = s.notifier.Delete(ctx, notifID)
		if err != nil {
			return fmt.Errorf("delete[notifID=%v]: %w", notifID, err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("tr: %w", err)
	}

	return nil
}

func (s *Service) DeleteBasicEvent(ctx context.Context, userID, eventID int) error {
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		event, err := s.repo.Events().Get(ctx, eventID)
		if err != nil {
			return fmt.Errorf("get event: %w", err)
		}

		if !event.BelongsTo(userID) {
			return serverrors.NewNotFoundError(err, "basic event")
		}

		notif, err := s.repo.Notifications().GetLatest(ctx, eventID)
		if err != nil {
			return fmt.Errorf("get latest notification: %w", err)
		}

		err = s.repo.Notifications().Delete(ctx, notif.ID)
		if err != nil {
			return fmt.Errorf("delete notification: %w", err)
		}

		err = s.repo.Events().Delete(ctx, eventID)
		if err != nil {
			return fmt.Errorf("delete basic event: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("tr: %w", err)
	}

	return nil
}
