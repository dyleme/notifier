package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Dyleme/Notifier/internal/timetable-service/repository/queries"
	"github.com/Dyleme/Notifier/internal/timetable-service/service"
)

type Repository struct {
	db    *pgxpool.Pool
	q     *queries.Queries
	cache Cache
}

type Cache interface {
	Get(key string, obj any) error
	Delete(key string) error
	Add(key string, obj any) error
}

func New(pool *pgxpool.Pool, cache Cache) *Repository {
	return &Repository{db: pool, q: queries.New(pool), cache: cache}
}

func (r *Repository) WithTx(tx pgx.Tx) *Repository {
	return &Repository{q: r.q.WithTx(tx), db: nil, cache: r.cache}
}

func (r *Repository) Atomic(ctx context.Context, fn func(ctx context.Context, repository service.Repository) error) error {
	op := "Repository.Atomic: %w"
	if r.db == nil {
		return fmt.Errorf(op, fmt.Errorf("cannot start transaction from another transaction"))
	}
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{}) //nolint:exhaustruct // default value for transactions
	if err != nil {
		return fmt.Errorf(op, err)
	}
	if err := fn(ctx, r.WithTx(tx)); err != nil {
		if rollErr := tx.Rollback(ctx); rollErr != nil {
			return fmt.Errorf(op, fmt.Errorf("rolling back transaction %w, (original error %w)", rollErr, err))
		}

		return fmt.Errorf(op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf(op, fmt.Errorf("committing transaction: %w", err))
	}

	return nil
}
