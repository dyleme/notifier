package repository

import (
	trmpgx "github.com/avito-tech/go-transaction-manager/pgxv5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Dyleme/Notifier/internal/service/repository/queries/goqueries"
)

type Repository struct {
	db    *pgxpool.Pool
	cache Cache

	periodicTaskRepository *PeriodicTaskRepository
	tasksRepository        *BasicTaskRepository
	eventParamsRepository  *NotificationParamsRepository
	tgImagesRepository     *TgImagesRepository
	eventsRepository       *EventsRepository
}

type Cache interface {
	Get(key string, obj any) error
	Delete(key string) error
	Add(key string, obj any) error
}

func New(pool *pgxpool.Pool, cache Cache, getter *trmpgx.CtxGetter) *Repository {
	q := goqueries.New()

	return &Repository{
		db:    pool,
		cache: cache,
		periodicTaskRepository: &PeriodicTaskRepository{
			q:      q,
			db:     pool,
			getter: getter,
		},
		tasksRepository: &BasicTaskRepository{
			q:      q,
			db:     pool,
			getter: getter,
		},
		eventParamsRepository: &NotificationParamsRepository{
			q:      q,
			getter: getter,
			db:     pool,
		},
		tgImagesRepository: &TgImagesRepository{
			q:      q,
			cache:  cache,
			getter: getter,
			db:     pool,
		},
		eventsRepository: &EventsRepository{
			q:      q,
			getter: getter,
			db:     pool,
		},
	}
}
