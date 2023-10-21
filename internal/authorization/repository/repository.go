package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Dyleme/Notifier/internal/authorization/repository/queries"
)

type Repository struct {
	db *pgxpool.Pool
	q  *queries.Queries
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db, q: queries.New(db)}
}
