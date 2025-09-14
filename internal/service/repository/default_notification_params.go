package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/internal/domain/apperr"
	"github.com/Dyleme/Notifier/internal/service/repository/queries/goqueries"
	"github.com/Dyleme/Notifier/pkg/database/txmanager"
)

type NotificationParamsRepository struct {
	q      *goqueries.Queries
	getter *txmanager.Getter
}

func NewDefaultNotificationParamsRepository(getter *txmanager.Getter) *NotificationParamsRepository {
	return &NotificationParamsRepository{
		getter: getter,
	}
}

func (nr *NotificationParamsRepository) Get(ctx context.Context, userID int) (domain.NotificationParams, error) {
	op := "eventParamsRepository.Get: %w"
	tx := nr.getter.GetTx(ctx)
	params, err := nr.q.GetDefaultUserNotificationParams(ctx, tx, int64(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NotificationParams{}, fmt.Errorf(op, apperr.NotFoundError{Object: "default eveft params"})
		}

		return domain.NotificationParams{}, fmt.Errorf(op, err)
	}

	return params.Params, nil
}

func (nr *NotificationParamsRepository) Set(ctx context.Context, userID int, params domain.NotificationParams) (domain.NotificationParams, error) {
	op := "eventParamsRepository.Set: %w"
	tx := nr.getter.GetTx(ctx)
	updatedParams, err := nr.q.SetDefaultUserNotificationParams(ctx, tx, goqueries.SetDefaultUserNotificationParamsParams{
		UserID: int64(userID),
		Params: params,
	})
	if err != nil {
		return domain.NotificationParams{}, fmt.Errorf(op, err)
	}

	return updatedParams.Params, nil
}
