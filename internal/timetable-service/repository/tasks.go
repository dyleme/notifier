package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/Dyleme/Notifier/internal/lib/serverrors"
	"github.com/Dyleme/Notifier/internal/lib/sql/pgxconv"
	"github.com/Dyleme/Notifier/internal/lib/utils/dto"
	"github.com/Dyleme/Notifier/internal/timetable-service/models"
	"github.com/Dyleme/Notifier/internal/timetable-service/repository/queries"
	"github.com/Dyleme/Notifier/internal/timetable-service/service"
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

type TaskRepository struct {
	q *queries.Queries
}

func (r *Repository) Tasks() service.TaskRepository {
	return &TaskRepository{q: r.q}
}

func (tr *TaskRepository) Get(ctx context.Context, id, userID int) (models.Task, error) {
	op := fmt.Sprintf("get task with (id{%v} userID{%v}): %%w", id, userID)
	task, err := tr.q.GetTask(ctx, queries.GetTaskParams{
		ID:     int32(id),
		UserID: int32(userID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Task{}, fmt.Errorf(op, serverrors.NewNotFoundError(err, "task"))
		}
		return models.Task{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}
	return dtoTask(task), nil
}

func (tr *TaskRepository) Add(ctx context.Context, task models.Task) (models.Task, error) {
	op := "add task: %%w"
	addedTask, err := tr.q.AddTask(ctx, queries.AddTaskParams{
		UserID:       int32(task.UserID),
		RequiredTime: pgxconv.Interval(task.RequiredTime),
		Message:      task.Text,
	})
	if err != nil {
		return models.Task{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}
	return dtoTask(addedTask), nil
}

func (tr *TaskRepository) Update(ctx context.Context, task models.Task) error {
	op := "update task: %%w"
	err := tr.q.UpdateTask(ctx, queries.UpdateTaskParams{
		ID:           int32(task.ID),
		UserID:       int32(task.UserID),
		RequiredTime: pgxconv.Interval(task.RequiredTime),
		Message:      task.Text,
		Periodic:     task.Periodic,
		Done:         task.Done,
		Archived:     task.Archived,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf(op, serverrors.NewNotFoundError(err, "task"))
		}
		return fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return nil
}

func (tr *TaskRepository) List(ctx context.Context, userID int) ([]models.Task, error) {
	op := fmt.Sprintf("list tasks userID{%v}: %%w", userID)
	tasks, err := tr.q.ListUserTasks(ctx, int32(userID))
	if err != nil {
		return nil, err
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf(op, serverrors.NewNotFoundError(err, "task"))
		}
		return nil, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return dto.Slice(tasks, dtoTask), nil
}

func (tr *TaskRepository) Delete(ctx context.Context, taskID, userID int) error {
	op := fmt.Sprintf("delete task with (id{%v} userID{%v}): %%w", taskID, userID)
	amount, err := tr.q.DeleteTask(ctx, queries.DeleteTaskParams{
		ID:     int32(taskID),
		UserID: int32(userID),
	})
	if err != nil {
		return fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}
	if amount == 0 {
		return fmt.Errorf(op, serverrors.NewNoDeletionsError("task"))
	}

	return nil
}
