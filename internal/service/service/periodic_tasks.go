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

func (s *Service) CreatePeriodicTask(ctx context.Context, perTask domains.PeriodicTask, userID int) (domains.PeriodicTask, error) {
	if err := perTask.BelongsTo(userID); err != nil {
		return domains.PeriodicTask{}, fmt.Errorf("belongs to: %w", serverrors.NewBusinessLogicError(err.Error()))
	}

	var createdPerTask domains.PeriodicTask
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		createdPerTask, err := s.repos.periodicTasks.Add(ctx, perTask)
		if err != nil {
			return fmt.Errorf("add periodic task: %w", err)
		}

		err = s.createAndAddEvent(ctx, createdPerTask, userID)
		if err != nil {
			return fmt.Errorf("create and add event: %w", err)
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
		pt, err = s.repos.periodicTasks.Get(ctx, taskID)
		if err != nil {
			return fmt.Errorf("get[taskID=%v]: %w", taskID, err)
		}

		if err := pt.BelongsTo(userID); err != nil {
			return fmt.Errorf("belongs to: %w", serverrors.NewBusinessLogicError(err.Error()))
		}

		return nil
	})
	if err != nil {
		return domains.PeriodicTask{}, fmt.Errorf("tr: %w", err)
	}

	return pt, nil
}

type UpdatePeriodicTaskParams struct {
	ID             int
	Text           string
	Description    string
	UserID         int
	Start          time.Duration // Event time from beginning of day
	SmallestPeriod time.Duration
	BiggestPeriod  time.Duration
	EventParams    *domains.NotificationParams
}

func (s *Service) UpdatePeriodicTask(ctx context.Context, perTask domains.PeriodicTask, userID int) error {
	var updatedTask domains.PeriodicTask
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var oldTask domains.PeriodicTask
		oldTask, err := s.repos.periodicTasks.Get(ctx, perTask.ID)
		if err != nil {
			return fmt.Errorf("get[taskID=%v]: %w", perTask.ID, err)
		}

		if err := oldTask.BelongsTo(userID); err != nil {
			return fmt.Errorf("belongs to: %w", serverrors.NewBusinessLogicError(err.Error()))
		}

		updatedTask, err = s.repos.periodicTasks.Update(ctx, perTask)
		if err != nil {
			return fmt.Errorf("update: %w", err)
		}

		event, err := s.repos.events.GetLatest(ctx, oldTask.ID, domains.PeriodicTaskType)
		if err != nil {
			return fmt.Errorf("delete event: %w", err)
		}

		if !updatedTask.TimeParamsHasChanged(oldTask) {
			event.Description = updatedTask.Description
			event.Text = updatedTask.Text

			err = s.repos.events.Update(ctx, event)
			if err != nil {
				return fmt.Errorf("update event: %w", err)
			}

			return nil
		}

		err = s.repos.events.Delete(ctx, event.ID)
		if err != nil {
			return fmt.Errorf("delete event: %w", err)
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
		task, err := s.repos.periodicTasks.Get(ctx, taskID)
		if err != nil {
			return fmt.Errorf("get current event: %w", err)
		}

		if err := task.BelongsTo(userID); err != nil {
			return fmt.Errorf("belongs to: %w", serverrors.NewBusinessLogicError(err.Error()))
		}

		event, err := s.repos.events.GetLatest(ctx, taskID, domains.PeriodicTaskType)
		if err != nil {
			return fmt.Errorf("get latest event: %w", err)
		}

		err = s.repos.events.Delete(ctx, event.ID)
		if err != nil {
			return fmt.Errorf("delete event: %w", err)
		}

		err = s.repos.periodicTasks.Delete(ctx, taskID)
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
	tasks, err := s.repos.periodicTasks.List(ctx, userID, listParams)
	if err != nil {
		err = fmt.Errorf("list periodic tasks: %w", err)
		logError(ctx, err)

		return nil, err
	}

	return tasks, nil
}

func (s *Service) createNewEventForPeriodicTask(ctx context.Context, taskID, userID int) error {
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		task, err := s.repos.periodicTasks.Get(ctx, taskID)
		if err != nil {
			return fmt.Errorf("periodic tasks get[taskID=%v,userID=%v]: %w", taskID, userID, err)
		}

		if err := task.BelongsTo(userID); err != nil {
			return fmt.Errorf("belongs to: %w", serverrors.NewBusinessLogicError(err.Error()))
		}

		err = s.createAndAddEvent(ctx, task, userID)
		if err != nil {
			return fmt.Errorf("create and add event: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("tr: %w", err)
	}

	return nil
}
