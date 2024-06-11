package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/pkg/serverrors"
)

//go:generate mockgen -destination=mocks/periodic_tasks_mocks.go -package=mocks . PeriodicTasksRepository
type PeriodicTasksRepository interface {
	Add(ctx context.Context, task domains.PeriodicTask) (domains.PeriodicTask, error)
	Get(ctx context.Context, taskID int) (domains.PeriodicTask, error)
	Update(ctx context.Context, task domains.PeriodicTask) (domains.PeriodicTask, error)
	Delete(ctx context.Context, taskID int) error
	List(ctx context.Context, userID int, listParams ListParams) ([]domains.PeriodicTask, error)
}

func (s *Service) AddPeriodicTask(ctx context.Context, perTask domains.PeriodicTask, userID int) (domains.PeriodicTask, error) {
	if !perTask.BelongsTo(userID) {
		return domains.PeriodicTask{}, fmt.Errorf("belongs to: %w", serverrors.NewBusinessLogicError("you are not allowed to add task to another user"))
	}

	var createdPerTask domains.PeriodicTask
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		createdPerTask, err := s.repo.PeriodicTasks().Add(ctx, perTask)
		if err != nil {
			return fmt.Errorf("add periodic task: %w", err)
		}

		notification, err := createdPerTask.NewNotification(time.Now())
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
		return domains.PeriodicTask{}, fmt.Errorf("tr: %w", err)
	}

	return createdPerTask, nil
}

func (s *Service) GetPeriodicTask(ctx context.Context, taskID, userID int) (domains.PeriodicTask, error) {
	var pt domains.PeriodicTask
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var err error
		pt, err = s.repo.PeriodicTasks().Get(ctx, taskID)
		if err != nil {
			return fmt.Errorf("get[taskID=%v]: %w", taskID, err)
		}

		if !pt.BelongsTo(userID) {
			return serverrors.NewNotFoundError(err, "periodic task")
		}

		return nil
	})
	if err != nil {
		return domains.PeriodicTask{}, fmt.Errorf("tr: %w", err)
	}

	return pt, nil
}

type UpdatePeriodicTaskParams struct {
	ID                 int
	Text               string
	Description        string
	UserID             int
	Start              time.Duration // Notification time from beginning of day
	SmallestPeriod     time.Duration
	BiggestPeriod      time.Duration
	NotificationParams *domains.NotificationParams
}

func (s *Service) UpdatePeriodicTask(ctx context.Context, perTask domains.PeriodicTask, userID int) error {
	var updatedTask domains.PeriodicTask
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var pt domains.PeriodicTask
		pt, err := s.repo.PeriodicTasks().Get(ctx, perTask.ID)
		if err != nil {
			return fmt.Errorf("get[taskID=%v]: %w", perTask.ID, err)
		}
		if !pt.BelongsTo(userID) {
			return serverrors.NewNotFoundError(err, "periodic task")
		}

		updatedTask, err = s.repo.PeriodicTasks().Update(ctx, perTask)
		if err != nil {
			return fmt.Errorf("update: %w", err)
		}

		notif, err := s.repo.Notifications().GetLatest(ctx, pt.ID)
		if err != nil {
			return fmt.Errorf("delete notification: %w", err)
		}

		nextNotif, err := updatedTask.NewNotification(time.Now())
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

func (s *Service) DeletePeriodicTask(ctx context.Context, taskID, userID int) error {
	op := "Service.DeletePeriodicTask: %w"

	err := s.tr.Do(ctx, func(ctx context.Context) error {
		task, err := s.repo.PeriodicTasks().Get(ctx, taskID)
		if err != nil {
			return fmt.Errorf("get current notifications: %w", err)
		}

		if task.BelongsTo(userID) {
			return serverrors.NewNotFoundError(err, "periodic task")
		}

		notif, err := s.repo.Notifications().GetLatest(ctx, taskID)
		if err != nil {
			return fmt.Errorf("get latest notification: %w", err)
		}

		err = s.repo.Notifications().Delete(ctx, notif.ID)
		if err != nil {
			return fmt.Errorf("delete notification: %w", err)
		}

		err = s.repo.PeriodicTasks().Delete(ctx, taskID)
		if err != nil {
			return fmt.Errorf("delete periodic task: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (s *Service) ListPeriodicTasks(ctx context.Context, userID int, listParams ListParams) ([]domains.PeriodicTask, error) {
	tasks, err := s.repo.PeriodicTasks().List(ctx, userID, listParams)
	if err != nil {
		err = fmt.Errorf("list periodic tasks: %w", err)
		logError(ctx, err)

		return nil, err
	}

	return tasks, nil
}
