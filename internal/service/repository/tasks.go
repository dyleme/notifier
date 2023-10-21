package repository

import (
	"context"
	"errors"
	"fmt"

	pgxgtm "github.com/avito-tech/go-transaction-manager/pgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/internal/service/repository/queries"
	"github.com/Dyleme/Notifier/internal/service/service"
	"github.com/Dyleme/Notifier/pkg/serverrors"
	"github.com/Dyleme/Notifier/pkg/utils"
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
	q      *queries.Queries
	getter *pgxgtm.CtxGetter
	db     *pgxpool.Pool
}

func (r *Repository) Tasks() service.TaskRepository {
	return r.taskRepository
}

func (tr *TaskRepository) Get(ctx context.Context, id, userID int) (domains.Task, error) {
	op := fmt.Sprintf("get task with (id{%v} userID{%v}): %%w", id, userID)
	tx := tr.getter.DefaultTrOrDB(ctx, tr.db)
	task, err := tr.q.GetTask(ctx, tx, queries.GetTaskParams{
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
	tx := tr.getter.DefaultTrOrDB(ctx, tr.db)
	addedTask, err := tr.q.AddTask(ctx, tx, queries.AddTaskParams{
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
	tx := tr.getter.DefaultTrOrDB(ctx, tr.db)
	err := tr.q.UpdateTask(ctx, tx, queries.UpdateTaskParams{
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
	tx := tr.getter.DefaultTrOrDB(ctx, tr.db)
	tasks, err := tr.q.ListTasks(ctx, tx, queries.ListTasksParams{
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

	return utils.DtoSlice(tasks, dtoTask), nil
}

func (tr *TaskRepository) Delete(ctx context.Context, taskID, userID int) error {
	op := fmt.Sprintf("delete task with (id{%v} userID{%v}): %%w", taskID, userID)
	tx := tr.getter.DefaultTrOrDB(ctx, tr.db)
	amount, err := tr.q.DeleteTask(ctx, tx, queries.DeleteTaskParams{
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
