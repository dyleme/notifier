package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/dyleme/Notifier/internal/domain"
	"github.com/dyleme/Notifier/internal/domain/apperr"
	"github.com/dyleme/Notifier/internal/repository/queries/goqueries"
	"github.com/dyleme/Notifier/pkg/database/sqlconv"
	"github.com/dyleme/Notifier/pkg/utils/slice"
)

func (r *EventsRepository) dto(dbEv goqueries.Event) domain.Event {
	event := domain.Event{
		TaskID:             int(dbEv.TaskID),
		SendingID:          int(dbEv.SendingID),
		Done:               sqlconv.ToBool(dbEv.Done),
		OriginalSending:    dbEv.OriginalSending,
		NextSending:        dbEv.NextSending,
		Text:               dbEv.Text,
		Descriptions:       dbEv.Description,
		TgID:               int(dbEv.TaskID),
		NotificationPeriod: time.Duration(dbEv.NotificationRetryPeriodS) * time.Second,
	}

	return event
}

func (r *EventsRepository) dtoSending(dbSnd goqueries.Sending) domain.Sending {
	return domain.Sending{
		ID:              int(dbSnd.ID),
		CreatedAt:       dbSnd.CreatedAt,
		TaskID:          int(dbSnd.TaskID),
		Done:            sqlconv.ToBool(dbSnd.Done),
		OriginalSending: dbSnd.OriginalSending,
		NextSending:     dbSnd.NextSending,
	}
}

func (r *EventsRepository) AddSending(ctx context.Context, event domain.Sending) error {
	tx := r.getter.GetTx(ctx)

	_, err := r.q.AddSending(ctx, tx, goqueries.AddSendingParams{
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

func (r *EventsRepository) GetSending(ctx context.Context, id int) (domain.Sending, error) {
	tx := r.getter.GetTx(ctx)

	event, err := r.q.GetSendning(ctx, tx, int64(id))
	if err != nil {
		return domain.Sending{}, fmt.Errorf("get event: %w", err)
	}

	return r.dtoSending(event), nil
}

func (r *EventsRepository) GetLatestSending(ctx context.Context, taskdID int) (domain.Sending, error) {
	tx := r.getter.GetTx(ctx)
	event, err := r.q.GetLatestSending(ctx, tx, int64(taskdID))
	if err != nil {
		return domain.Sending{}, fmt.Errorf("get latest event[taskID=%d]: %w", taskdID, err)
	}

	return r.dtoSending(event), nil
}

func (r *EventsRepository) UpdateSending(ctx context.Context, event domain.Sending) error {
	tx := r.getter.GetTx(ctx)
	_, err := r.q.UpdateSending(ctx, tx, goqueries.UpdateSendingParams{
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

func (r *EventsRepository) DeleteSending(ctx context.Context, id int) error {
	tx := r.getter.GetTx(ctx)

	ns, err := r.q.DeleteSending(ctx, tx, int64(id))
	if err != nil {
		return fmt.Errorf("delete event: %w", err)
	}

	if len(ns) == 0 {
		return fmt.Errorf("delete event: %w", apperr.ErrNotFound)
	}

	return nil
}

func (r *EventsRepository) ListNotSended(ctx context.Context, till time.Time) ([]domain.Event, error) {
	tx := r.getter.GetTx(ctx)

	dbEvents, err := r.q.ListNotSentEvents(ctx, tx, till)
	if err != nil {
		return nil, fmt.Errorf("list not sended notifiations: %w", err)
	}

	return slice.Dto(dbEvents, r.dto), nil
}
