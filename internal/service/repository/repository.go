package repository

import (
	trmpgx "github.com/avito-tech/go-transaction-manager/pgxv5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Dyleme/Notifier/internal/service/repository/queries"
)

type Repository struct {
	db    *pgxpool.Pool
	cache Cache

	periodicEventRepository      *PeriodicEventRepository
	eventsRepository             *EventRepository
	taskRepository               *TaskRepository
	notificationParamsRepository *NotificationParamsRepository
	tgImagesRepository           *TgImagesRepository
}

type Cache interface {
	Get(key string, obj any) error
	Delete(key string) error
	Add(key string, obj any) error
}

func New(pool *pgxpool.Pool, cache Cache, getter *trmpgx.CtxGetter) *Repository {
	q := queries.New()

	return &Repository{
		db:    pool,
		cache: cache,
		periodicEventRepository: &PeriodicEventRepository{
			q:      q,
			db:     pool,
			getter: getter,
		},
		eventsRepository: &EventRepository{
			q:      q,
			db:     pool,
			getter: getter,
		},
		taskRepository: &TaskRepository{
			q:      q,
			getter: getter,
			db:     pool,
		},
		notificationParamsRepository: &NotificationParamsRepository{
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
	}
}
