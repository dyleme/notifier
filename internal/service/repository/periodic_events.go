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
	"github.com/Dyleme/Notifier/internal/service/repository/queries"
	"github.com/Dyleme/Notifier/internal/service/service"
	"github.com/Dyleme/Notifier/pkg/serverrors"
	"github.com/Dyleme/Notifier/pkg/sql/pgxconv"
	"github.com/Dyleme/Notifier/pkg/utils"
)

type PeriodicEventRepository struct {
	q      *queries.Queries
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func (r *Repository) PeriodicEvents() service.PeriodicEventsRepository {
	return r.periodicEventRepository
}

func (p *PeriodicEventRepository) dto(ev queries.PeriodicEvent) domains.PeriodicEvent {
	return domains.PeriodicEvent{
		ID:                 int(ev.ID),
		Text:               ev.Text,
		Description:        pgxconv.String(ev.Description),
		UserID:             int(ev.UserID),
		Start:              pgxconv.TimeWithZone(ev.Start).Sub(time.Time{}),
		SmallestPeriod:     time.Duration(ev.SmallestPeriod) * time.Minute,
		BiggestPeriod:      time.Duration(ev.BiggestPeriod) * time.Minute,
		NotificationParams: ev.NotificationParams,
	}
}

func (p *PeriodicEventRepository) Add(ctx context.Context, event domains.PeriodicEvent) (domains.PeriodicEvent, error) {
	tx := p.getter.DefaultTrOrDB(ctx, p.db)
	ev, err := p.q.AddPeriodicEvent(ctx, tx, queries.AddPeriodicEventParams{
		UserID:             int32(event.UserID),
		Text:               event.Text,
		Start:              pgxconv.Timestamptz(time.Time{}.Add(event.Start)),
		SmallestPeriod:     int32(event.SmallestPeriod / time.Minute),
		BiggestPeriod:      int32(event.BiggestPeriod / time.Minute),
		Description:        pgxconv.Text(event.Description),
		NotificationParams: event.NotificationParams,
	})
	if err != nil {
		return domains.PeriodicEvent{}, fmt.Errorf("add periodic event: %w", serverrors.NewRepositoryError(err))
	}

	return p.dto(ev), nil
}

func (p *PeriodicEventRepository) Get(ctx context.Context, eventID int) (domains.PeriodicEvent, error) {
	op := "PeriodicEventRepository.Get: %w"

	tx := p.getter.DefaultTrOrDB(ctx, p.db)
	ev, err := p.q.GetPeriodicEvent(ctx, tx, int32(eventID))
	if err != nil {
		return domains.PeriodicEvent{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return p.dto(ev), nil
}

func (p *PeriodicEventRepository) Update(ctx context.Context, event domains.PeriodicEvent) (domains.PeriodicEvent, error) {
	op := "PeriodicEventRepository.Update: %w"

	tx := p.getter.DefaultTrOrDB(ctx, p.db)
	ev, err := p.q.UpdatePeriodicEvent(ctx, tx, queries.UpdatePeriodicEventParams{
		ID:                 int32(event.ID),
		UserID:             int32(event.UserID),
		Start:              pgxconv.Timestamptz(time.Time{}.Add(event.Start)),
		Text:               event.Text,
		Description:        pgxconv.Text(event.Description),
		NotificationParams: event.NotificationParams,
		SmallestPeriod:     int32(event.SmallestPeriod / time.Minute),
		BiggestPeriod:      int32(event.BiggestPeriod / time.Minute),
	})
	if err != nil {
		return domains.PeriodicEvent{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return p.dto(ev), nil
}

func (p *PeriodicEventRepository) Delete(ctx context.Context, eventID int) error {
	op := "PeriodicEventRepository.Delete: %w"

	tx := p.getter.DefaultTrOrDB(ctx, p.db)
	evs, err := p.q.DeletePeriodicEvent(ctx, tx, int32(eventID))
	if err != nil {
		return fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}
	if len(evs) == 0 {
		return fmt.Errorf(op, serverrors.NewNoDeletionsError("periodic event"))
	}

	return nil
}

func (p *PeriodicEventRepository) List(ctx context.Context, userID int, listParams service.ListParams) ([]domains.PeriodicEvent, error) {
	tx := p.getter.DefaultTrOrDB(ctx, p.db)

	events, err := p.q.ListPeriodicEvents(ctx, tx, queries.ListPeriodicEventsParams{
		UserID: int32(userID),
		Off:    int32(listParams.Offset),
		Lim:    int32(listParams.Limit),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("list periodic events: %w", serverrors.NewRepositoryError(err))
	}

	return utils.DtoSlice(events, p.dto), nil
}
