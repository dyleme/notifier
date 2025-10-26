package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dyleme/Notifier/internal/domain"
	"github.com/dyleme/Notifier/internal/domain/apperr"
	"github.com/dyleme/Notifier/pkg/log"
)

//go:generate mockgen -destination=mocks/basic_tasks_mocks.go -package=mocks . SingleTaskRepository
type TaskRepository interface {
	Add(ctx context.Context, task domain.Task) (domain.Task, error)
	List(ctx context.Context, userID int, taskType domain.TaskType, params ListParams) ([]domain.Task, error)
	Update(ctx context.Context, task domain.Task) error
	Delete(ctx context.Context, taskID, userID int) error
	Get(ctx context.Context, taskID, userID int) (domain.Task, error)
}

func (s *Service) addTask(ctx context.Context, task domain.Task, sending domain.Sending) error {
	task, err := s.repos.tasks.Add(ctx, task)
	if err != nil {
		return fmt.Errorf("add task: %w", err)
	}

	sending.TaskID = task.ID
	err = s.repos.events.AddSending(ctx, sending)
	if err != nil {
		return fmt.Errorf("add event: %w", err)
	}

	return nil
}

func (s *Service) updateTask(ctx context.Context, task domain.Task, sending domain.Sending) error {
	err := s.repos.tasks.Update(ctx, task)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	oldSending, err := s.repos.events.GetLatestSending(ctx, task.ID)
	if err != nil {
		return fmt.Errorf("get latest event[taskID=%d]: %w", task.ID, err)
	}

	sending.ID = oldSending.ID

	err = s.repos.events.UpdateSending(ctx, sending)
	if err != nil {
		return fmt.Errorf("update event: %w", err)
	}

	return nil
}

func (s *Service) DeleteTask(ctx context.Context, userID, taskID int) error {
	log.Ctx(ctx).Debug("delete task", "userID", userID, "taskID", taskID)
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		sending, err := s.repos.events.GetLatestSending(ctx, taskID)
		if err == nil {
			// delete existed event
			err = s.repos.events.DeleteSending(ctx, sending.ID)
			if err != nil {
				return fmt.Errorf("delete event: %w", err)
			}
		}

		if !errors.Is(err, apperr.ErrNotFound) {
			return fmt.Errorf("get latest event[taskID=%d]: %w", taskID, err)
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

func (s *Service) createNewSending(ctx context.Context, taskID, userID int) error {
	task, err := s.repos.tasks.Get(ctx, taskID, userID)
	if err != nil {
		return fmt.Errorf("get task: %w", err)
	}
	var sending domain.Sending

	switch task.Type {
	case domain.Single:
		return nil
	case domain.Periodic:
		pt, err := domain.ParsePeriodicTask(task)
		if err != nil {
			return fmt.Errorf("parse periodic task: %w", err)
		}
		sending = pt.NewSending(time.Now())
	default:
		return fmt.Errorf("unknown task type: %v", task.Type)
	}

	log.Ctx(ctx).Debug("new sending", "sending", sending, "task", task)

	err = s.repos.events.AddSending(ctx, sending)
	if err != nil {
		return fmt.Errorf("add sending: %w", err)
	}

	return nil
}
