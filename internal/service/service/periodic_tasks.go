package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/internal/domain/apperr"
	"github.com/Dyleme/Notifier/pkg/log"
)

//go:generate mockgen -destination=mocks/periodic_tasks_mocks.go -package=mocks . PeriodicTasksRepository
type PeriodicTasksRepository interface {
	Add(ctx context.Context, task domain.PeriodicTask) (domain.PeriodicTask, error)
	Get(ctx context.Context, taskID int) (domain.PeriodicTask, error)
	Update(ctx context.Context, task domain.PeriodicTask) error
	Delete(ctx context.Context, taskID int) error
	List(ctx context.Context, userID int, params ListFilterParams) ([]domain.PeriodicTask, error)
}

func (s *Service) CreatePeriodicTask(ctx context.Context, perTask domain.PeriodicTask, userID int) (domain.PeriodicTask, error) {
	log.Ctx(ctx).Debug("creating periodic task", slog.Any("task", perTask), slog.Int("userID", userID))
	if err := perTask.BelongsTo(userID); err != nil {
		return domain.PeriodicTask{}, err
	}

	var createdPerTask domain.PeriodicTask
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
		return domain.PeriodicTask{}, fmt.Errorf("tr: %w", err)
	}

	return createdPerTask, nil
}

func (s *Service) GetPeriodicTask(ctx context.Context, taskID, userID int) (domain.PeriodicTask, error) {
	log.Ctx(ctx).Debug("getting periodic task", "taskID", taskID, "userID", userID)
	var pt domain.PeriodicTask
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var err error
		pt, err = s.repos.periodicTasks.Get(ctx, taskID)
		if err != nil {
			return fmt.Errorf("get[taskID=%v]: %w", taskID, err)
		}

		if err := pt.BelongsTo(userID); err != nil {
			return fmt.Errorf("belongs to: %w", err)
		}

		return nil
	})
	if err != nil {
		return domain.PeriodicTask{}, fmt.Errorf("tr: %w", err)
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
	EventParams    *domain.NotificationParams
}

func (s *Service) UpdatePeriodicTask(ctx context.Context, perTask domain.PeriodicTask, userID int) error {
	log.Ctx(ctx).Debug("updating periodic task", "task", perTask, "userID", userID)
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var oldTask domain.PeriodicTask
		oldTask, err := s.repos.periodicTasks.Get(ctx, perTask.ID)
		if err != nil {
			return fmt.Errorf("get[taskID=%v]: %w", perTask.ID, err)
		}

		if err := oldTask.BelongsTo(userID); err != nil {
			return fmt.Errorf("belongs to: %w", err)
		}

		err = s.repos.periodicTasks.Update(ctx, perTask)
		if err != nil {
			return fmt.Errorf("update: %w", err)
		}

		event, err := s.repos.events.GetLatest(ctx, oldTask.ID, domain.PeriodicTaskType)
		if err != nil {
			return fmt.Errorf("delete event: %w", err)
		}

		updatedTask, err := s.repos.periodicTasks.Get(ctx, perTask.ID)
		if err != nil {
			return fmt.Errorf("get[taskID=%v]: %w", perTask.ID, err)
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
	log.Ctx(ctx).Debug("deleting periodic task", "taskID", taskID, "userID", userID)
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		task, err := s.repos.periodicTasks.Get(ctx, taskID)
		if err != nil {
			return fmt.Errorf("get current event: %w", err)
		}

		if err := task.BelongsTo(userID); err != nil {
			return fmt.Errorf("belongs to: %w", err)
		}

		event, err := s.repos.events.GetLatest(ctx, taskID, domain.PeriodicTaskType)
		if err != nil {
			return fmt.Errorf("get latest event: %w", err)
		}

		err = s.repos.events.Delete(ctx, event.ID)
		if err != nil {
			return fmt.Errorf("delete event: %w", err)
		}

		err = s.repos.periodicTasks.Delete(ctx, taskID)
		if err != nil {
			if errors.Is(err, apperr.ErrNotFound) {
				return apperr.NotFoundError{Object: "periodic task"}
			}

			return fmt.Errorf("delete periodic task: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("tr: %w", err)
	}

	return nil
}

func (s *Service) ListPeriodicTasks(ctx context.Context, userID int, params ListFilterParams) ([]domain.PeriodicTask, error) {
	log.Ctx(ctx).Debug("list periodic tasks", "userID", userID, "listparams", params)
	var tasks []domain.PeriodicTask
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var err error
		tasks, err = s.repos.periodicTasks.List(ctx, userID, params)
		if err != nil {
			return fmt.Errorf("list periodic tasks: %w", err)
		}

		return nil
	})
	if err != nil {
		err = fmt.Errorf("tr: %w", err)

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
			return fmt.Errorf("belongs to: %w", err)
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
