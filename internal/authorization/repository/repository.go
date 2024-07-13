package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Dyleme/Notifier/internal/authorization/repository/queries/goqueries"
)

type Repository struct {
	db *pgxpool.Pool
	q  *goqueries.Queries
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db, q: goqueries.New(db)}
}
