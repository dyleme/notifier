package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/pkg/utils/timeborders"
)

//go:generate mockgen -destination=mocks/events_mocks.go -package=mocks . EventsRepository
type EventsRepository interface {
	Add(ctx context.Context, event domains.Event) (domains.Event, error)
	List(ctx context.Context, userID int, timeBorderes timeborders.TimeBorders, listParams ListParams) ([]domains.Event, error)
	Get(ctx context.Context, id int) (domains.Event, error)
	GetLatest(ctx context.Context, taskdID int) (domains.Event, error)
	Update(ctx context.Context, event domains.Event) error
	Delete(ctx context.Context, id int) error
	ListNotSended(ctx context.Context, till time.Time) ([]domains.Event, error)
	GetNearest(ctx context.Context, till time.Time) (domains.Event, error)
	MarkSended(ctx context.Context, ids []int) error
}

func (s *Service) ListEvents(ctx context.Context, userID int, timeBorders timeborders.TimeBorders, listParams ListParams) ([]domains.Event, error) {
	var events []domains.Event
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var err error
		events, err = s.repo.Events().List(ctx, userID, timeBorders, listParams)
		if err != nil {
			return fmt.Errorf("events: list: %w", err)
		}

		return nil
	})
	if err != nil {
		logError(ctx, err)

		return nil, fmt.Errorf("tr: %w", err)
	}

	return events, nil
}
