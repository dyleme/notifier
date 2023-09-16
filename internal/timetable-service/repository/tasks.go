package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/Dyleme/Notifier/internal/lib/serverrors"
	"github.com/Dyleme/Notifier/internal/lib/utils/dto"
	"github.com/Dyleme/Notifier/internal/timetable-service/domains"
	"github.com/Dyleme/Notifier/internal/timetable-service/repository/queries"
	"github.com/Dyleme/Notifier/internal/timetable-service/service"
)

func dtoTask(task queries.Task) domains.Task {
	return domains.Task{
		ID:       int(task.ID),
		UserID:   int(task.UserID),
		Text:     task.Message,
		Periodic: task.Periodic,
		Archived: task.Archived,
	}
}

type TaskRepository struct {
	q *queries.Queries
}

func (r *Repository) Tasks() service.TaskRepository {
	return &TaskRepository{q: r.q}
}

func (tr *TaskRepository) Get(ctx context.Context, id, userID int) (domains.Task, error) {
	op := fmt.Sprintf("get task with (id{%v} userID{%v}): %%w", id, userID)
	task, err := tr.q.GetTask(ctx, queries.GetTaskParams{
		ID:     int32(id),
		UserID: int32(userID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domains.Task{}, fmt.Errorf(op, serverrors.NewNotFoundError(err, "task"))
		}

		return domains.Task{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return dtoTask(task), nil
}

func (tr *TaskRepository) Add(ctx context.Context, task domains.Task) (domains.Task, error) {
	op := "add task: %%w"
	addedTask, err := tr.q.AddTask(ctx, queries.AddTaskParams{
		UserID:   int32(task.UserID),
		Message:  task.Text,
		Periodic: task.Periodic,
	})
	if err != nil {
		return domains.Task{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return dtoTask(addedTask), nil
}

func (tr *TaskRepository) Update(ctx context.Context, task domains.Task) error {
	op := "update task: %w"
	err := tr.q.UpdateTask(ctx, queries.UpdateTaskParams{
		ID:       int32(task.ID),
		UserID:   int32(task.UserID),
		Message:  task.Text,
		Periodic: task.Periodic,
		Archived: task.Archived,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf(op, serverrors.NewNotFoundError(err, "task"))
		}

		return fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return nil
}

func (tr *TaskRepository) List(ctx context.Context, userID int, listParams service.ListParams) ([]domains.Task, error) {
	op := fmt.Sprintf("list tasks userID{%v}: %%w", userID)
	tasks, err := tr.q.ListTasks(ctx, queries.ListTasksParams{
		UserID: int32(userID),
		Off:    int32(listParams.Offset),
		Lim:    int32(listParams.Limit),
	})
	if err != nil {
		return nil, fmt.Errorf(op, err)
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
