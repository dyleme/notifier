package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/internal/domain/apperr"
	"github.com/Dyleme/Notifier/internal/repository/queries/goqueries"
	"github.com/Dyleme/Notifier/internal/service"
	"github.com/Dyleme/Notifier/pkg/database/txmanager"
	"github.com/Dyleme/Notifier/pkg/log"
	"github.com/Dyleme/Notifier/pkg/utils/slice"
)

type EventsRepository struct {
	q      *goqueries.Queries
	getter *txmanager.Getter
}

func NewEventsRepository(getter *txmanager.Getter) *EventsRepository {
	return &EventsRepository{
		getter: getter,
		q:      &goqueries.Queries{},
	}
}

func (r *EventsRepository) Get(ctx context.Context, sendingID, userID int) (domain.Event, error) {
	tx := r.getter.GetTx(ctx)
	dbEvent, err := r.q.GetEvent(ctx, tx,
		goqueries.GetEventParams{
			SendingID: int64(sendingID),
			UserID:    int64(userID),
		})
	if err != nil {
		return domain.Event{}, err
	}

	return r.dto(dbEvent), nil
}

func (r *EventsRepository) List(ctx context.Context, userID int, params service.ListEventsFilterParams) ([]domain.Event, error) {
	tx := r.getter.GetTx(ctx)
	log.Ctx(ctx).Debug("args", slog.Any("arg", params))
	dbEvents, err := r.q.ListEvents(ctx, tx, goqueries.ListEventsParams{
		UserID:   int64(userID),
		Limit:    int64(params.ListParams.Limit),
		Offset:   int64(params.ListParams.Offset),
		FromTime: params.TimeBorders.From,
		ToTime:   params.TimeBorders.To,
	})
	if err != nil {
		return nil, err
	}

	return slice.Dto(dbEvents, r.dto), nil
}

func (r *EventsRepository) ListNotSent(
	ctx context.Context, till time.Time,
) ([]domain.Event, error) {
	tx := r.getter.GetTx(ctx)
	dbNotifications, err := r.q.ListNotSentEvents(ctx, tx, till)
	if err != nil {
		return nil, err
	}

	notifs := slice.Dto(dbNotifications,
		func(e goqueries.Event) domain.Event {
			return domain.Event{
				SendingID:          int(e.SendingID),
				NextSending:        e.NextSending,
				Text:               e.Text,
				TgID:               int(e.TgID),
				NotificationPeriod: time.Duration(e.NotificationRetryPeriodS) * time.Second,
				TaskID:             int(e.TaskID),
				Descriptions:       e.Description,
			}
		},
	)

	return notifs, nil
}

func (r *EventsRepository) GetNearest(ctx context.Context) (time.Time, error) {
	tx := r.getter.GetTx(ctx)

	t, err := r.q.GetNearestSendingTime(ctx, tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return time.Time{}, fmt.Errorf("get nearest event: %w", apperr.ErrNotFound)
		}

		return time.Time{}, fmt.Errorf("list not sended notifiations: %w", err)
	}

	return t, nil
}
