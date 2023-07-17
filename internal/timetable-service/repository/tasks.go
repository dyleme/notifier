package repository

import (
	"context"
	"fmt"

	"github.com/Dyleme/Notifier/internal/lib/sql/pgxconv"
	"github.com/Dyleme/Notifier/internal/timetable-service/models"
	"github.com/Dyleme/Notifier/internal/timetable-service/repository/queries"
)

func dtoTask(task queries.Task) models.Task {
	return models.Task{
		ID:           int(task.ID),
		UserID:       int(task.UserID),
		RequiredTime: pgxconv.Duration(task.RequiredTime),
		Text:         task.Message,
		Periodic:     task.Periodic,
		Done:         task.Done,
		Archived:     task.Archived,
	}
}

func dtoTasks(tasks []queries.Task) []models.Task {
	tsks := make([]models.Task, 0, len(tasks))
	for _, t := range tasks {
		tsks = append(tsks, dtoTask(t))
	}
	return tsks
}

func (r *Repository) GetTask(ctx context.Context, id, userID int) (models.Task, error) {
	task, err := r.q.GetTask(ctx, queries.GetTaskParams{
		ID:     int32(id),
		UserID: int32(userID),
	})
	if err != nil {
		return models.Task{}, nil //nolint:exhaustruct // nil value
	}
	return dtoTask(task), nil
}

func (r *Repository) AddTask(ctx context.Context, task models.Task) (models.Task, error) {
	addedTask, err := r.q.AddTask(ctx, queries.AddTaskParams{
		UserID:       int32(task.UserID),
		RequiredTime: pgxconv.Interval(task.RequiredTime),
		Message:      task.Text,
	})
	if err != nil {
		return models.Task{}, err
	}
	return dtoTask(addedTask), nil
}

func (r *Repository) UpdateTask(ctx context.Context, task models.Task) error {
	return r.q.UpdateTask(ctx, queries.UpdateTaskParams{
		ID:           int32(task.ID),
		UserID:       int32(task.UserID),
		RequiredTime: pgxconv.Interval(task.RequiredTime),
		Message:      task.Text,
		Periodic:     task.Periodic,
		Done:         task.Done,
		Archived:     task.Archived,
	})
}

func (r *Repository) ListTasks(ctx context.Context, userID int) ([]models.Task, error) {
	tasks, err := r.q.ListUserTasks(ctx, int32(userID))
	if err != nil {
		return nil, err
	}

	return dtoTasks(tasks), nil
}

func (r *Repository) DeleteTask(ctx context.Context, taskID, userID int) error {
	amount, err := r.q.DeleteTask(ctx, queries.DeleteTaskParams{
		ID:     int32(taskID),
		UserID: int32(userID),
	})
	if err != nil {
		return err
	}
	if amount == 0 {
		return fmt.Errorf("no rows deleted")
	}

	return nil
}
