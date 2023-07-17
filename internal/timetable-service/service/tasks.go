package service

import (
	"context"

	"github.com/Dyleme/Notifier/internal/timetable-service/models"
)

func (s *Service) AddTask(ctx context.Context, task models.Task) (models.Task, error) {
	createdTask, err := s.repo.AddTask(ctx, task)
	if err != nil {
		return models.Task{}, err
	}

	return createdTask, nil
}

func (s *Service) GetTask(ctx context.Context, taskID, userID int) (models.Task, error) {
	task, err := s.repo.GetTask(ctx, taskID, userID)
	if err != nil {
		return models.Task{}, err
	}

	return task, nil
}

func (s *Service) UpdateTask(ctx context.Context, task models.Task) error {
	return s.repo.UpdateTask(ctx, task)
}

func (s *Service) GetUserTasks(ctx context.Context, userID int) ([]models.Task, error) {
	tasks, err := s.repo.ListTasks(ctx, userID)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}
