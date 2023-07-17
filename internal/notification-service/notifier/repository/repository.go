package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Dyleme/Notifier/internal/notification-service/notifier/repository/queries"
)

type Repository struct {
	db *sql.DB
	q  *queries.Queries
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db, q: queries.New(db)}
}

var defaulTxOpts = &sql.TxOptions{
	Isolation: sql.LevelDefault,
	ReadOnly:  false,
}

func (r *Repository) inTx(ctx context.Context, txOpts *sql.TxOptions, fn func(q *queries.Queries) error) error {
	tx, err := r.db.BeginTx(ctx, txOpts)
	if err != nil {
		return err
	}

	if err := fn(r.q.WithTx(tx)); err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return fmt.Errorf("rolling back transaction %w, (original error %w)", rollErr, err)
		}

		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}
