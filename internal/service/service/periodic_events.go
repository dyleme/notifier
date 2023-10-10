package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/pkg/serverrors"
)

//go:generate mockgen -destination=mocks/periodic_events_mocks.go -package=mocks . PeriodicEventsRepository
type PeriodicEventsRepository interface {
	Add(context.Context, domains.PeriodicEvent) (domains.PeriodicEvent, error)
	Get(ctx context.Context, eventID, userID int) (domains.PeriodicEvent, error)
	Update(ctx context.Context, event domains.PeriodicEvent) (domains.PeriodicEvent, error)
	Delete(ctx context.Context, eventID, userID int) error
	AddNotification(context.Context, domains.PeriodicEventNotification) (domains.PeriodicEventNotification, error)
	UpdateNotification(context.Context, domains.PeriodicEventNotification) error
	GetCurrentNotification(ctx context.Context, eventID int) (domains.PeriodicEventNotification, error)
	DeleteNotification(ctx context.Context, notifID, eventID int) error
	GetNearestNotificationSendTime(ctx context.Context) (time.Time, error)
	ListNotificationsAtSendTime(ctx context.Context, sendTime time.Time) ([]domains.PeriodicEventWithNotification, error)
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
	perEvent.ID, err = equalOrChangeIfZero(perEvent.ID, userID)
	if err != nil {
		return domains.PeriodicEvent{}, fmt.Errorf(op, err)
	}

	var createdPerEvent domains.PeriodicEvent
	err = s.repo.Atomic(ctx, func(ctx context.Context, repo Repository) error {
		createdPerEvent, err = repo.PeriodicEvents().Add(ctx, perEvent)
		if err != nil {
			return fmt.Errorf("add: %w", err)
		}

		notification := createdPerEvent.NextNotification()
		_, err = repo.PeriodicEvents().AddNotification(ctx, notification)
		if err != nil {
			return fmt.Errorf("add notification: %w", err)
		}

		return nil
	})
	if err != nil {
		return domains.PeriodicEvent{}, fmt.Errorf(op, err)
	}

	return createdPerEvent, nil
}

func (s *Service) DonePeriodicEvent(ctx context.Context, eventID, userID int) error {
	op := "Service.DonePeriodicEvent: %w"

	err := s.repo.Atomic(ctx, func(ctx context.Context, repo Repository) error {
		event, err := repo.PeriodicEvents().Get(ctx, eventID, userID)
		if err != nil {
			return fmt.Errorf("get: %w", err)
		}

		notif, err := repo.PeriodicEvents().GetCurrentNotification(ctx, event.ID)
		if err != nil {
			return fmt.Errorf("get current notification: %w", err)
		}

		notif.Done = true
		err = repo.PeriodicEvents().UpdateNotification(ctx, notif)
		if err != nil {
			return fmt.Errorf("update notification: %w", err)
		}
		nextNotif := event.NextNotification()
		_, err = repo.PeriodicEvents().AddNotification(ctx, nextNotif)
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

func (s *Service) UpdatePeriodicEvent(ctx context.Context, perEvent domains.PeriodicEvent, userID int) error {
	op := "Service.UpdatePeriodicEvent: %w"

	var err error
	perEvent.UserID, err = equalOrChangeIfZero(perEvent.UserID, userID)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	var updatedEvent domains.PeriodicEvent
	err = s.repo.Atomic(ctx, func(ctx context.Context, repo Repository) error {
		var err error //nolint:govet // return this error
		updatedEvent, err = repo.PeriodicEvents().Update(ctx, perEvent)
		if err != nil {
			return fmt.Errorf("update: %w", err)
		}

		currNotif, err := repo.PeriodicEvents().GetCurrentNotification(ctx, perEvent.ID)
		if err != nil {
			return fmt.Errorf("get current notification: %w", err)
		}

		err = repo.PeriodicEvents().DeleteNotification(ctx, currNotif.ID, perEvent.ID)
		if err != nil {
			return fmt.Errorf("delete notification: %w", err)
		}

		nextNotif := updatedEvent.NextNotification()
		_, err = repo.PeriodicEvents().AddNotification(ctx, nextNotif)
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

func (s *Service) DeletePeriodicEvent(ctx context.Context, eventID, userID int) error {
	op := "Service.DeletePeriodicEvent: %w"

	err := s.repo.Atomic(ctx, func(ctx context.Context, repo Repository) error {
		notif, err := repo.PeriodicEvents().GetCurrentNotification(ctx, eventID)
		if err != nil {
			return fmt.Errorf("get current notifications: %w", err)
		}

		err = repo.PeriodicEvents().DeleteNotification(ctx, notif.ID, eventID)
		if err != nil {
			return fmt.Errorf("delete notification: %w", err)
		}

		err = repo.PeriodicEvents().Delete(ctx, eventID, userID)
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
