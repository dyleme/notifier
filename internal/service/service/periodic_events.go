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
	Add(ctx context.Context, event domains.PeriodicEvent) (domains.PeriodicEvent, error)
	Get(ctx context.Context, eventID int) (domains.PeriodicEvent, error)
	Update(ctx context.Context, event domains.PeriodicEvent) (domains.PeriodicEvent, error)
	Delete(ctx context.Context, eventID int) error
	List(ctx context.Context, userID int, listParams ListParams) ([]domains.PeriodicEvent, error)
}

func (s *Service) AddPeriodicEvent(ctx context.Context, perEvent domains.PeriodicEvent, userID int) (domains.PeriodicEvent, error) {
	if !perEvent.BelongsTo(userID) {
		return domains.PeriodicEvent{}, serverrors.NewBusinessLogicError("you are not allowed to add event to another user")
	}

	var createdPerEvent domains.PeriodicEvent
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		createdPerEvent, err := s.repo.PeriodicEvents().Add(ctx, perEvent)
		if err != nil {
			return fmt.Errorf("add periodic event: %w", err)
		}

		notification, err := createdPerEvent.NewNotification(time.Now())
		if err != nil {
			return fmt.Errorf("next notification: %w", serverrors.NewBusinessLogicError(err.Error()))
		}

		_, err = s.repo.Notifications().Add(ctx, notification)
		if err != nil {
			return fmt.Errorf("add notification: %w", err)
		}

		return nil
	})
	if err != nil {
		return domains.PeriodicEvent{}, fmt.Errorf("tr: %w", err)
	}

	return createdPerEvent, nil
}

func (s *Service) GetPeriodicEvent(ctx context.Context, eventID, userID int) (domains.PeriodicEvent, error) {
	var ev domains.PeriodicEvent
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var err error
		ev, err = s.repo.PeriodicEvents().Get(ctx, eventID)
		if err != nil {
			return fmt.Errorf("get[eventID=%v]: %w", eventID, err)
		}

		if !ev.BelongsTo(userID) {
			return serverrors.NewNotFoundError(err, "periodic event")
		}

		return nil
	})
	if err != nil {
		return domains.PeriodicEvent{}, fmt.Errorf("tr: %w", err)
	}

	return ev, nil
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

func (s *Service) UpdatePeriodicEvent(ctx context.Context, perEvent domains.PeriodicEvent, userID int) error {
	var updatedEvent domains.PeriodicEvent
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var ev domains.PeriodicEvent
		ev, err := s.repo.PeriodicEvents().Get(ctx, perEvent.ID)
		if err != nil {
			return fmt.Errorf("get[eventID=%v]: %w", perEvent.ID, err)
		}
		if !ev.BelongsTo(userID) {
			return serverrors.NewNotFoundError(err, "periodic event")
		}

		updatedEvent, err = s.repo.PeriodicEvents().Update(ctx, perEvent)
		if err != nil {
			return fmt.Errorf("update: %w", err)
		}

		notif, err := s.repo.Notifications().GetLatest(ctx, ev.ID)
		if err != nil {
			return fmt.Errorf("delete notification: %w", err)
		}

		nextNotif, err := updatedEvent.NewNotification(time.Now())
		if err != nil {
			return fmt.Errorf("next notification: %w", serverrors.NewBusinessLogicError(err.Error()))
		}

		nextNotif.ID = notif.ID

		err = s.repo.Notifications().Update(ctx, nextNotif)
		if err != nil {
			return fmt.Errorf("add notification: %w", err)
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
		event, err := s.repo.PeriodicEvents().Get(ctx, eventID)
		if err != nil {
			return fmt.Errorf("get current notifications: %w", err)
		}

		if event.BelongsTo(userID) {
			return serverrors.NewNotFoundError(err, "periodic event")
		}

		notif, err := s.repo.Notifications().GetLatest(ctx, eventID)
		if err != nil {
			return fmt.Errorf("get latest notification: %w", err)
		}

		err = s.repo.Notifications().Delete(ctx, notif.ID)
		if err != nil {
			return fmt.Errorf("delete notification: %w", err)
		}

		err = s.repo.PeriodicEvents().Delete(ctx, eventID)
		if err != nil {
			return fmt.Errorf("delete periodic event: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (s *Service) ListPeriodicEvents(ctx context.Context, userID int, listParams ListParams) ([]domains.PeriodicEvent, error) {
	events, err := s.repo.PeriodicEvents().List(ctx, userID, listParams)
	if err != nil {
		err = fmt.Errorf("list periodic events: %w", err)
		logError(ctx, err)

		return nil, err
	}

	return events, nil
}
