package repository

import (
	trmpgx "github.com/avito-tech/go-transaction-manager/pgxv5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Dyleme/Notifier/internal/service/repository/queries"
)

type Repository struct {
	db    *pgxpool.Pool
	cache Cache

	periodicTaskRepository       *PeriodicTaskRepository
	tasksRepository              *BasicTaskRepository
	notificationParamsRepository *NotificationParamsRepository
	tgImagesRepository           *TgImagesRepository
	notificationsRepository      *NotificationsRepository
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
		notificationsRepository: &NotificationsRepository{
			q:      q,
			getter: getter,
			db:     pool,
		},
	}
}
