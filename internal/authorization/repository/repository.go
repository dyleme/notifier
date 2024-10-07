package repository

import (
	trmpgx "github.com/avito-tech/go-transaction-manager/pgxv5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Dyleme/Notifier/internal/authorization/repository/queries/goqueries"
)

type Repository struct {
	q      *goqueries.Queries
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func New(pool *pgxpool.Pool, getter *trmpgx.CtxGetter) *Repository {
	return &Repository{
		q:      &goqueries.Queries{},
		db:     pool,
		getter: getter,
	}
}
