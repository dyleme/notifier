package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/internal/domain/apperr"
	"github.com/Dyleme/Notifier/pkg/log"
)

//go:generate mockgen -destination=mocks/basic_tasks_mocks.go -package=mocks . SingleTaskRepository
type TaskRepository interface {
	Add(ctx context.Context, task domain.Task) (domain.Task, error)
	List(ctx context.Context, userID int, params ListParams) ([]domain.Task, error)
	Update(ctx context.Context, task domain.Task) error
	Delete(ctx context.Context, taskID, userID int) error
	Get(ctx context.Context, taskID, userID int) (domain.Task, error)
}

func (s *Service) addTask(ctx context.Context, task domain.Task, event domain.Sending) error {
	_, err := s.repos.tasks.Add(ctx, task)
	if err != nil {
		return fmt.Errorf("add task: %w", err)
	}

	err = s.repos.events.Add(ctx, event)
	if err != nil {
		return fmt.Errorf("add event: %w", err)
	}

	return nil
}

func (s *Service) updateTask(ctx context.Context, task domain.Task, event domain.Sending) error {
	err := s.repos.tasks.Update(ctx, task)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	oldEvent, err := s.repos.events.GetLatest(ctx, task.ID)
	if err != nil {
		return fmt.Errorf("get latest event: %w", err)
	}

	event.ID = oldEvent.ID

	err = s.repos.events.Update(ctx, event)
	if err != nil {
		return fmt.Errorf("update event: %w", err)
	}

	return nil
}

func (s *Service) DeleteTask(ctx context.Context, userID, taskID int) error {
	log.Ctx(ctx).Debug("delete task", "userID", userID, "taskID", taskID)
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		event, err := s.repos.events.GetLatest(ctx, taskID)
		if err != nil {
			return fmt.Errorf("get latest event: %w", err)
		}

		err = s.repos.events.Delete(ctx, event.ID)
		if err != nil {
			if errors.Is(err, apperr.ErrNotFound) {
				return apperr.NotFoundError{Object: "event"}
			}

			return fmt.Errorf("delete event: %w", err)
		}

		err = s.repos.tasks.Delete(ctx, taskID, userID)
		if err != nil {
			if errors.Is(err, apperr.ErrNotFound) {
				return apperr.NotFoundError{Object: "task"}
			}

			return fmt.Errorf("delete single task: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("tr: %w", err)
	}

	return nil
}

func (s *Service) createNewEvent(ctx context.Context, taskID, userID int) error {
	task, err := s.repos.tasks.Get(ctx, taskID, userID)
	if err != nil {
		return fmt.Errorf("get task: %w", err)
	}

	var event domain.Sending

	switch task.Type {
	case domain.Signle:
		return nil
	case domain.Periodic:
		perTask, err := domain.PeriodictaskFromTask(task)
		if err != nil {
			return fmt.Errorf("periodic task from task: %w", err)
		}

		event = perTask.NewEvent(time.Now())

	default:
		return fmt.Errorf("unknown task type: %v", task.Type)
	}

	err = s.repos.events.Add(ctx, event)
	if err != nil {
		return fmt.Errorf("add event: %w", err)
	}

	return nil
}
