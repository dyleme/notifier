package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/internal/domain/apperr"
	"github.com/Dyleme/Notifier/internal/repository/queries/goqueries"
	"github.com/Dyleme/Notifier/internal/service"
	"github.com/Dyleme/Notifier/pkg/database/sqlconv"
	"github.com/Dyleme/Notifier/pkg/database/txmanager"
	"github.com/Dyleme/Notifier/pkg/utils/slice"
)

type SendingRepository struct {
	q      *goqueries.Queries
	getter *txmanager.Getter
}

func NewSendingRepository(getter *txmanager.Getter) *SendingRepository {
	return &SendingRepository{
		getter: getter,
		q:      &goqueries.Queries{},
	}
}

func (er *SendingRepository) dto(dbEv goqueries.Sending) (domain.Sending, error) {
	event := domain.Sending{
		ID:              int(dbEv.ID),
		CreatedAt:       dbEv.CreatedAt,
		TaskID:          int(dbEv.TaskID),
		Done:            sqlconv.ToBool(dbEv.Done),
		OriginalSending: dbEv.OriginalSending,
		NextSending:     dbEv.NextSending,
	}

	return event, nil
}

func (er *SendingRepository) Add(ctx context.Context, event domain.Sending) error {
	tx := er.getter.GetTx(ctx)

	_, err := er.q.AddSending(ctx, tx, goqueries.AddSendingParams{
		TaskID:          int64(event.TaskID),
		Done:            sqlconv.BoolToInt(event.Done),
		OriginalSending: event.OriginalSending,
		NextSending:     event.NextSending,
	})
	if err != nil {
		return fmt.Errorf("add event: %w", err)
	}

	return nil
}

func (er *SendingRepository) List(ctx context.Context, userID int, params service.ListEventsFilterParams) ([]domain.Sending, error) {
	tx := er.getter.GetTx(ctx)

	rowsSendings, err := er.q.ListUserSending(ctx, tx, goqueries.ListUserSendingParams{
		UserID:   int64(userID),
		FromTime: params.TimeBorders.From,
		ToTime:   params.TimeBorders.To,
		Offset:   int64(params.ListParams.Offset),
		Limit:    int64(params.ListParams.Limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list user events: %w", err)
	}

	events, err := slice.DtoError(rowsSendings, er.dto)
	if err != nil {
		return nil, fmt.Errorf("list user events: %w", err)
	}

	return events, nil
}

func (er *SendingRepository) Get(ctx context.Context, id int) (domain.Sending, error) {
	tx := er.getter.GetTx(ctx)

	event, err := er.q.GetSendning(ctx, tx, int64(id))
	if err != nil {
		return domain.Sending{}, fmt.Errorf("get event: %w", err)
	}

	return er.dto(event)
}

func (er *SendingRepository) GetLatest(ctx context.Context, taskdID int) (domain.Sending, error) {
	tx := er.getter.GetTx(ctx)
	event, err := er.q.GetLatestSending(ctx, tx, int64(taskdID))
	if err != nil {
		return domain.Sending{}, fmt.Errorf("get latest event: %w", err)
	}

	return er.dto(event)
}

func (er *SendingRepository) Update(ctx context.Context, event domain.Sending) error {
	tx := er.getter.GetTx(ctx)
	_, err := er.q.UpdateSending(ctx, tx, goqueries.UpdateSendingParams{
		NextSending:     event.NextSending,
		OriginalSending: event.OriginalSending,
		Done:            sqlconv.BoolToInt(event.Done),
		ID:              int64(event.ID),
	})
	if err != nil {
		return fmt.Errorf("update event: %w", err)
	}

	return nil
}

func (er *SendingRepository) Delete(ctx context.Context, id int) error {
	tx := er.getter.GetTx(ctx)

	ns, err := er.q.DeleteSending(ctx, tx, int64(id))
	if err != nil {
		return fmt.Errorf("delete event: %w", err)
	}

	if len(ns) == 0 {
		return fmt.Errorf("delete event: %w", apperr.ErrNotFound)
	}

	return nil
}

func (er *SendingRepository) ListNotSended(ctx context.Context, till time.Time) ([]domain.Sending, error) {
	tx := er.getter.GetTx(ctx)

	dbSendings, err := er.q.ListNotSendedSending(ctx, tx, till)
	if err != nil {
		return nil, fmt.Errorf("list not sended notifiations: %w", err)
	}

	events, err := slice.DtoError(dbSendings, er.dto)
	if err != nil {
		return nil, fmt.Errorf("list not sended notifiations: %w", err)
	}

	return events, nil
}

func (er *SendingRepository) GetNearest(ctx context.Context) (time.Time, error) {
	tx := er.getter.GetTx(ctx)

	t, err := er.q.GetNearestSendingTime(ctx, tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return time.Time{}, fmt.Errorf("get nearest event: %w", apperr.ErrNotFound)
		}

		return time.Time{}, fmt.Errorf("list not sended notifiations: %w", err)
	}

	return t, nil
}
