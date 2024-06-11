package repository

import (
	"context"
	"errors"
	"fmt"

	trmpgx "github.com/avito-tech/go-transaction-manager/pgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/internal/service/repository/queries"
	"github.com/Dyleme/Notifier/internal/service/service"
	"github.com/Dyleme/Notifier/pkg/serverrors"
	"github.com/Dyleme/Notifier/pkg/sql/pgxconv"
	"github.com/Dyleme/Notifier/pkg/utils"
)

type BasicTaskRepository struct {
	q      *queries.Queries
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func (r *Repository) Tasks() service.BasicTaskRepository {
	return r.tasksRepository
}

func dtoTask(bt queries.BasicTask) (domains.BasicTask, error) {
	return domains.BasicTask{
		ID:                 int(bt.ID),
		UserID:             int(bt.UserID),
		Text:               bt.Text,
		Description:        pgxconv.String(bt.Description),
		Start:              pgxconv.TimeWithZone(bt.Start),
		NotificationParams: bt.NotificationParams,
	}, nil
}

func (er *BasicTaskRepository) Add(ctx context.Context, bt domains.BasicTask) (domains.BasicTask, error) {
	op := "add timetable task: %w"

	tx := er.getter.DefaultTrOrDB(ctx, er.db)
	addedTask, err := er.q.AddBasicTask(ctx, tx, queries.AddBasicTaskParams{
		UserID:             int32(bt.UserID),
		Text:               bt.Text,
		Start:              pgxconv.Timestamptz(bt.Start),
		Description:        pgxconv.Text(bt.Description),
		NotificationParams: bt.NotificationParams,
	})
	if err != nil {
		return domains.BasicTask{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return dtoTask(addedTask)
}

func (er *BasicTaskRepository) List(ctx context.Context, userID int, listParams service.ListParams) ([]domains.BasicTask, error) {
	op := fmt.Sprintf("list timetable tasks userID{%v} %%w", userID)
	tx := er.getter.DefaultTrOrDB(ctx, er.db)
	tt, err := er.q.ListBasicTasks(ctx, tx, queries.ListBasicTasksParams{
		UserID: int32(userID),
		Off:    int32(listParams.Offset),
		Lim:    int32(listParams.Limit),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	tasks, err := utils.DtoErrorSlice(tt, dtoTask)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return tasks, nil
}

func (er *BasicTaskRepository) Delete(ctx context.Context, taskID int) error {
	tx := er.getter.DefaultTrOrDB(ctx, er.db)
	deletedTasks, err := er.q.DeleteBasicTask(ctx, tx, int32(taskID))
	if err != nil {
		return fmt.Errorf("delete basic task: %w", serverrors.NewRepositoryError(err))
	}
	if len(deletedTasks) == 0 {
		return serverrors.NewNoDeletionsError("tasks")
	}

	return nil
}

func (er *BasicTaskRepository) Get(ctx context.Context, taskID int) (domains.BasicTask, error) {
	tx := er.getter.DefaultTrOrDB(ctx, er.db)
	tt, err := er.q.GetBasicTask(ctx, tx, int32(taskID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domains.BasicTask{}, fmt.Errorf("get basic task: %w", serverrors.NewNotFoundError(err, "basic task"))
		}

		return domains.BasicTask{}, fmt.Errorf("get basic task: %w", serverrors.NewRepositoryError(err))
	}

	return dtoTask(tt)
}

func (er *BasicTaskRepository) Update(ctx context.Context, task domains.BasicTask) (domains.BasicTask, error) {
	op := "update timetable task: %w"
	tx := er.getter.DefaultTrOrDB(ctx, er.db)
	updatedTT, err := er.q.UpdateBasicTask(ctx, tx, queries.UpdateBasicTaskParams{
		Start:              pgxconv.Timestamptz(task.Start),
		Text:               task.Text,
		Description:        pgxconv.Text(task.Description),
		NotificationParams: task.NotificationParams,
		ID:                 int32(task.ID),
		UserID:             int32(task.UserID),
	})
	if err != nil {
		return domains.BasicTask{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return dtoTask(updatedTT)
}
