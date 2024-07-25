package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/pgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/internal/service/repository/queries/goqueries"
	"github.com/Dyleme/Notifier/internal/service/service"
	"github.com/Dyleme/Notifier/pkg/serverrors"
	"github.com/Dyleme/Notifier/pkg/sql/pgxconv"
	"github.com/Dyleme/Notifier/pkg/utils"
)

type PeriodicTaskRepository struct {
	q      *goqueries.Queries
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewPeriodicTaskRepository(db *pgxpool.Pool, getter *trmpgx.CtxGetter) *PeriodicTaskRepository {
	return &PeriodicTaskRepository{
		q:      goqueries.New(),
		db:     db,
		getter: getter,
	}
}

func (p *PeriodicTaskRepository) dto(pt goqueries.PeriodicTask) domains.PeriodicTask {
	return domains.PeriodicTask{
		ID:                 int(pt.ID),
		Text:               pt.Text,
		Description:        pgxconv.String(pt.Description),
		UserID:             int(pt.UserID),
		Start:              pgxconv.TimeWithZone(pt.Start).Sub(time.Time{}),
		SmallestPeriod:     time.Duration(pt.SmallestPeriod) * time.Minute,
		BiggestPeriod:      time.Duration(pt.BiggestPeriod) * time.Minute,
		NotificationParams: pt.NotificationParams,
	}
}

func (p *PeriodicTaskRepository) Add(ctx context.Context, task domains.PeriodicTask) (domains.PeriodicTask, error) {
	tx := p.getter.DefaultTrOrDB(ctx, p.db)
	pt, err := p.q.AddPeriodicTask(ctx, tx, goqueries.AddPeriodicTaskParams{
		UserID:             int32(task.UserID),
		Text:               task.Text,
		Start:              pgxconv.Timestamptz(time.Time{}.Add(task.Start)),
		SmallestPeriod:     int32(task.SmallestPeriod / time.Minute),
		BiggestPeriod:      int32(task.BiggestPeriod / time.Minute),
		Description:        pgxconv.Text(task.Description),
		NotificationParams: task.NotificationParams,
	})
	if err != nil {
		return domains.PeriodicTask{}, fmt.Errorf("add periodic task: %w", serverrors.NewRepositoryError(err))
	}

	return p.dto(pt), nil
}

func (p *PeriodicTaskRepository) Get(ctx context.Context, taskID int) (domains.PeriodicTask, error) {
	op := "PeriodicTaskRepository.Get: %w"

	tx := p.getter.DefaultTrOrDB(ctx, p.db)
	pt, err := p.q.GetPeriodicTask(ctx, tx, int32(taskID))
	if err != nil {
		return domains.PeriodicTask{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return p.dto(pt), nil
}

func (p *PeriodicTaskRepository) Update(ctx context.Context, task domains.PeriodicTask) (domains.PeriodicTask, error) {
	op := "PeriodicTaskRepository.Update: %w"

	tx := p.getter.DefaultTrOrDB(ctx, p.db)
	pt, err := p.q.UpdatePeriodicTask(ctx, tx, goqueries.UpdatePeriodicTaskParams{
		ID:                 int32(task.ID),
		UserID:             int32(task.UserID),
		Start:              pgxconv.Timestamptz(time.Time{}.Add(task.Start)),
		Text:               task.Text,
		Description:        pgxconv.Text(task.Description),
		NotificationParams: task.NotificationParams,
		SmallestPeriod:     int32(task.SmallestPeriod / time.Minute),
		BiggestPeriod:      int32(task.BiggestPeriod / time.Minute),
	})
	if err != nil {
		return domains.PeriodicTask{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return p.dto(pt), nil
}

func (p *PeriodicTaskRepository) Delete(ctx context.Context, taskID int) error {
	op := "PeriodicTaskRepository.Delete: %w"

	tx := p.getter.DefaultTrOrDB(ctx, p.db)
	evs, err := p.q.DeletePeriodicTask(ctx, tx, int32(taskID))
	if err != nil {
		return fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}
	if len(evs) == 0 {
		return fmt.Errorf(op, serverrors.NewNoDeletionsError("periodic task"))
	}

	return nil
}

func (p *PeriodicTaskRepository) List(ctx context.Context, userID int, listParams service.ListParams) ([]domains.PeriodicTask, error) {
	tx := p.getter.DefaultTrOrDB(ctx, p.db)

	tasks, err := p.q.ListPeriodicTasks(ctx, tx, goqueries.ListPeriodicTasksParams{
		UserID: int32(userID),
		Off:    int32(listParams.Offset),
		Lim:    int32(listParams.Limit),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("list periodic tasks: %w", serverrors.NewRepositoryError(err))
	}

	return utils.DtoSlice(tasks, p.dto), nil
}
