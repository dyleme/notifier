package service

import (
	"context"
	"fmt"

	"github.com/Dyleme/Notifier/internal/timetable-service/models"
)

type TaskRepository interface {
	Add(ctx context.Context, task models.Task) (models.Task, error)
	Get(ctx context.Context, taskID, userID int) (models.Task, error)
	Delete(ctx context.Context, taskID, userID int) error
	Update(ctx context.Context, task models.Task) error
	List(ctx context.Context, userID int) ([]models.Task, error)
}

func (s *Service) AddTask(ctx context.Context, task models.Task) (models.Task, error) {
	op := "Service.AddTask: %w"
	createdTask, err := s.repo.Tasks().Add(ctx, task)
	if err != nil {
		logError(ctx, fmt.Errorf(op, err))
		return models.Task{}, err
	}

	return createdTask, nil
}

func (s *Service) GetTask(ctx context.Context, taskID, userID int) (models.Task, error) {
	op := "Service.GetTask: %w"
	task, err := s.repo.Tasks().Get(ctx, taskID, userID)
	if err != nil {
		logError(ctx, fmt.Errorf(op, err))
		return models.Task{}, err
	}

	return task, nil
}

func (s *Service) UpdateTask(ctx context.Context, task models.Task) error {
	op := "Service.UpdateTask: %w"
	err := s.repo.Tasks().Update(ctx, task)
	if err != nil {
		logError(ctx, fmt.Errorf(op, err))
		return err
	}

	return nil
}

func (s *Service) ListUserTasks(ctx context.Context, userID int) ([]models.Task, error) {
	op := "Service.ListUserTasks: %w"
	tasks, err := s.repo.Tasks().List(ctx, userID)
	if err != nil {
		logError(ctx, fmt.Errorf(op, err))
		return nil, err
	}

	return tasks, nil
}
