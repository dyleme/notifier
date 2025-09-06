package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Dyleme/Notifier/internal/authorization/repository/queries/goqueries"
	"github.com/Dyleme/Notifier/pkg/database/txmanager"
)

type Repository struct {
	q      *goqueries.Queries
	getter *txmanager.Getter
}

func New(pool *pgxpool.Pool, getter *txmanager.Getter) *Repository {
	return &Repository{
		q:      &goqueries.Queries{},
		getter: getter,
	}
}
