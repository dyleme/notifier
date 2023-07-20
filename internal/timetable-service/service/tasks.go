package service

import (
	"context"

	"github.com/Dyleme/Notifier/internal/timetable-service/models"
)

func (s *Service) AddTask(ctx context.Context, task models.Task) (models.Task, error) {
	createdTask, err := s.repo.AddTask(ctx, task)
	if err != nil {
		logError(ctx, err)
		return models.Task{}, err
	}

	return createdTask, nil
}

func (s *Service) GetTask(ctx context.Context, taskID, userID int) (models.Task, error) {
	task, err := s.repo.GetTask(ctx, taskID, userID)
	if err != nil {
		logError(ctx, err)
		return models.Task{}, err
	}

	return task, nil
}

func (s *Service) UpdateTask(ctx context.Context, task models.Task) error {
	err := s.repo.UpdateTask(ctx, task)
	if err != nil {
		logError(ctx, err)
		return err
	}

	return nil
}

func (s *Service) GetUserTasks(ctx context.Context, userID int) ([]models.Task, error) {
	tasks, err := s.repo.ListTasks(ctx, userID)
	if err != nil {
		logError(ctx, err)
		return nil, err
	}

	return tasks, nil
}
