package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/pkg/serverrors"
)

//go:generate mockgen -destination=mocks/periodic_events_mocks.go -package=mocks . PeriodicEventsRepository
type PeriodicEventsRepository interface { //nolint:interfacebloat // need so many interfaces
	Add(ctx context.Context, event domains.PeriodicEvent, notif domains.PeriodicEventNotification) (domains.PeriodicEvent, error)
	Get(ctx context.Context, eventID, userID int) (domains.PeriodicEvent, error)
	Update(ctx context.Context, event UpdatePeriodicEventParams) (domains.PeriodicEvent, error)
	Delete(ctx context.Context, eventID, userID int) error
	AddNotification(ctx context.Context, notif domains.PeriodicEventNotification) (domains.PeriodicEventNotification, error)
	MarkNotificationSend(ctx context.Context, notifID int) error
	MarkNotificationDone(ctx context.Context, eventID, userID int) error
	DeleteNotification(ctx context.Context, notifID, eventID int) error
	DeleteNotifications(ctx context.Context, eventID int) error
	GetNearestNotificationSendTime(ctx context.Context) (time.Time, error)
	ListNotificationsAtSendTime(ctx context.Context, sendTime time.Time) ([]domains.PeriodicEvent, error)
	ListFutureEvents(ctx context.Context, userID int, listParams ListParams) ([]domains.PeriodicEvent, error)
}

func equalOrChangeIfZero[T comparable](check, val T) (T, error) {
	if check == val {
		return check, nil
	}

	var zero T
	if check == zero {
		return val, nil
	}

	return check, serverrors.NewBusinessLogicError("not equal")
}

func (s *Service) AddPeriodicEvent(ctx context.Context, perEvent domains.PeriodicEvent, userID int) (domains.PeriodicEvent, error) {
	op := "Service.AddPeriodicEvent: %w"

	var err error
	perEvent.UserID, err = equalOrChangeIfZero(perEvent.UserID, userID)
	if err != nil {
		return domains.PeriodicEvent{}, fmt.Errorf("equal or change if zero[ev.UserID=%v,userID=%v]: %w", perEvent.UserID, userID, err)
	}

	var createdPerEvent domains.PeriodicEvent
	err = s.tr.Do(ctx, func(ctx context.Context) error {
		notification, err := perEvent.NextNotification()
		if err != nil {
			return fmt.Errorf("next notification: %w", serverrors.NewBusinessLogicError(err.Error()))
		}
		createdPerEvent, err = s.repo.PeriodicEvents().Add(ctx, perEvent, notification)
		if err != nil {
			return fmt.Errorf("add: %w", err)
		}

		return nil
	})
	if err != nil {
		return domains.PeriodicEvent{}, fmt.Errorf(op, err)
	}

	return createdPerEvent, nil
}

func (s *Service) GetPeriodicEvent(ctx context.Context, eventID, userID int) (domains.PeriodicEvent, error) {
	var ev domains.PeriodicEvent
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var err error
		ev, err = s.repo.PeriodicEvents().Get(ctx, eventID, userID)
		if err != nil {
			return fmt.Errorf("get[eventID=%v,userID=%v]: %w", eventID, userID, err)
		}

		return nil
	})
	if err != nil {
		return domains.PeriodicEvent{}, fmt.Errorf("atmoic: %w", err)
	}

	return ev, nil
}

func (s *Service) DonePeriodicEvent(ctx context.Context, eventID, userID int) error {
	op := "Service.DonePeriodicEvent: %w"

	err := s.tr.Do(ctx, func(ctx context.Context) error {
		event, err := s.repo.PeriodicEvents().Get(ctx, eventID, userID)
		if err != nil {
			return fmt.Errorf("get: %w", err)
		}

		err = s.repo.PeriodicEvents().MarkNotificationSend(ctx, event.Notification.ID)
		if err != nil {
			return fmt.Errorf("mark notified[notifID=%v]: %w", event.Notification.ID, err)
		}

		nextNotif, err := event.NextNotification()
		if err != nil {
			return fmt.Errorf("next notification: %w", serverrors.NewBusinessLogicError(err.Error()))
		}
		_, err = s.repo.PeriodicEvents().AddNotification(ctx, nextNotif)
		if err != nil {
			return fmt.Errorf("add notification: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

type UpdatePeriodicEventParams struct {
	ID                 int
	Text               string
	Description        string
	UserID             int
	Start              time.Duration // Notification time from beginning of day
	SmallestPeriod     time.Duration
	BiggestPeriod      time.Duration
	NotificationParams *domains.NotificationParams
}

func (s *Service) UpdatePeriodicEvent(ctx context.Context, perEvent UpdatePeriodicEventParams, userID int) error {
	var err error
	perEvent.UserID, err = equalOrChangeIfZero(perEvent.UserID, userID)
	if err != nil {
		return fmt.Errorf("equal or change if zero[ev.UserID=%v,userID=%v]: %w", perEvent.UserID, userID, err)
	}

	var updatedEvent domains.PeriodicEvent
	err = s.tr.Do(ctx, func(ctx context.Context) error {
		var ev domains.PeriodicEvent
		ev, err = s.repo.PeriodicEvents().Get(ctx, perEvent.ID, userID)
		if err != nil {
			return fmt.Errorf("get[eventID=%v]: %w", perEvent.ID, err)
		}

		updatedEvent, err = s.repo.PeriodicEvents().Update(ctx, perEvent)
		if err != nil {
			return fmt.Errorf("update: %w", err)
		}

		if ev.NeedRegenerateNotification(updatedEvent) {
			err = s.repo.PeriodicEvents().DeleteNotification(ctx, updatedEvent.Notification.ID, perEvent.ID)
			if err != nil {
				return fmt.Errorf("delete notification: %w", err)
			}

			nextNotif, err := updatedEvent.NextNotification()
			if err != nil {
				return fmt.Errorf("next notification: %w", serverrors.NewBusinessLogicError(err.Error()))
			}
			_, err = s.repo.PeriodicEvents().AddNotification(ctx, nextNotif)
			if err != nil {
				return fmt.Errorf("add notification: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("atomoic: %w", err)
	}

	return nil
}

func (s *Service) DeletePeriodicEvent(ctx context.Context, eventID, userID int) error {
	op := "Service.DeletePeriodicEvent: %w"

	err := s.tr.Do(ctx, func(ctx context.Context) error {
		event, err := s.repo.PeriodicEvents().Get(ctx, eventID, userID)
		if err != nil {
			return fmt.Errorf("get current notifications: %w", err)
		}

		err = s.repo.PeriodicEvents().DeleteNotifications(ctx, event.ID)
		if err != nil {
			return fmt.Errorf("delete notification: %w", err)
		}

		err = s.repo.PeriodicEvents().Delete(ctx, eventID, userID)
		if err != nil {
			return fmt.Errorf("delete: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (s *Service) ListFuturePeriodicEvents(ctx context.Context, userID int, listParams ListParams) ([]domains.PeriodicEvent, error) {
	op := "Service.ListEventsInPeriod: %w"
	evWithNotifs, err := s.repo.PeriodicEvents().ListFutureEvents(ctx, userID, listParams)
	if err != nil {
		err = fmt.Errorf(op, err)
		logError(ctx, err)

		return nil, err
	}

	return evWithNotifs, nil
}
