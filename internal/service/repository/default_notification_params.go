package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/internal/service/repository/queries"
	"github.com/Dyleme/Notifier/internal/service/service"
	serverrors2 "github.com/Dyleme/Notifier/pkg/serverrors"
)

type NotificationParamsRepository struct {
	q *queries.Queries
}

func (r *Repository) DefaultNotificationParams() service.NotificationParamsRepository {
	return &NotificationParamsRepository{q: r.q}
}

func (nr *NotificationParamsRepository) Get(ctx context.Context, userID int) (domains.NotificationParams, error) {
	op := "notificationParamsRepository.Get: %w"
	params, err := nr.q.GetDefaultUserNotificationsParams(ctx, int32(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domains.NotificationParams{}, fmt.Errorf(op, serverrors2.NewNotFoundError(err, "default notification params"))
		}

		return domains.NotificationParams{}, fmt.Errorf(op, serverrors2.NewRepositoryError(err))
	}

	return params.Params, nil
}

func (nr *NotificationParamsRepository) Set(ctx context.Context, userID int, params domains.NotificationParams) (domains.NotificationParams, error) {
	op := "notificationParamsRepository.Set: %w"
	updatedParams, err := nr.q.SetDefaultUserNotificationParams(ctx, queries.SetDefaultUserNotificationParamsParams{
		UserID: int32(userID),
		Params: params,
	})
	if err != nil {
		return domains.NotificationParams{}, fmt.Errorf(op, serverrors2.NewRepositoryError(err))
	}

	return updatedParams.Params, nil
}
