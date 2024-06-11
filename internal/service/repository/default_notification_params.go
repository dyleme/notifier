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
)

type NotificationParamsRepository struct {
	q      *queries.Queries
	getter *trmpgx.CtxGetter
	db     *pgxpool.Pool
}

func (r *Repository) DefaultEventParams() service.NotificationParamsRepository {
	return r.eventParamsRepository
}

func (nr *NotificationParamsRepository) Get(ctx context.Context, userID int) (domains.NotificationParams, error) {
	op := "eventParamsRepository.Get: %w"
	tx := nr.getter.DefaultTrOrDB(ctx, nr.db)
	params, err := nr.q.GetDefaultUserNotificationParams(ctx, tx, int32(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domains.NotificationParams{}, fmt.Errorf(op, serverrors.NewNotFoundError(err, "default event params"))
		}

		return domains.NotificationParams{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return params.Params, nil
}

func (nr *NotificationParamsRepository) Set(ctx context.Context, userID int, params domains.NotificationParams) (domains.NotificationParams, error) {
	op := "eventParamsRepository.Set: %w"
	tx := nr.getter.DefaultTrOrDB(ctx, nr.db)
	updatedParams, err := nr.q.SetDefaultUserNotificationParams(ctx, tx, queries.SetDefaultUserNotificationParamsParams{
		UserID: int32(userID),
		Params: params,
	})
	if err != nil {
		return domains.NotificationParams{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return updatedParams.Params, nil
}
