package repository

import (
	"context"
	"errors"
	"fmt"

	trmpgx "github.com/avito-tech/go-transaction-manager/pgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/internal/service/repository/queries/goqueries"
	"github.com/Dyleme/Notifier/pkg/serverrors"
)

type NotificationParamsRepository struct {
	q      *goqueries.Queries
	getter *trmpgx.CtxGetter
	db     *pgxpool.Pool
}

func NewDefaultNotificationParamsRepository(db *pgxpool.Pool, getter *trmpgx.CtxGetter) *NotificationParamsRepository {
	return &NotificationParamsRepository{
		q:      goqueries.New(),
		getter: getter,
		db:     db,
	}
}

func (nr *NotificationParamsRepository) Get(ctx context.Context, userID int) (domain.NotificationParams, error) {
	op := "eventParamsRepository.Get: %w"
	tx := nr.getter.DefaultTrOrDB(ctx, nr.db)
	params, err := nr.q.GetDefaultUserNotificationParams(ctx, tx, int32(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NotificationParams{}, fmt.Errorf(op, serverrors.NewNotFoundError(err, "default event params"))
		}

		return domain.NotificationParams{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return params.Params, nil
}

func (nr *NotificationParamsRepository) Set(ctx context.Context, userID int, params domain.NotificationParams) (domain.NotificationParams, error) {
	op := "eventParamsRepository.Set: %w"
	tx := nr.getter.DefaultTrOrDB(ctx, nr.db)
	updatedParams, err := nr.q.SetDefaultUserNotificationParams(ctx, tx, goqueries.SetDefaultUserNotificationParamsParams{
		UserID: int32(userID),
		Params: params,
	})
	if err != nil {
		return domain.NotificationParams{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return updatedParams.Params, nil
}
