package txmanager

import (
	"context"
	"database/sql"
	"fmt"
)

// TxManager manages database transactions.
type TxManager struct {
	db   *sql.DB
	opts []Option
}

func newTxManager(db *sql.DB, opts []Option) *TxManager {
	return &TxManager{
		db:   db,
		opts: opts,
	}
}

func (tm *TxManager) Do(ctx context.Context, fn func(context.Context) error) error {
	return tm.DoWithSettings(ctx, nil, fn)
}

// WithTransaction executes a function within a transaction.
func (tm *TxManager) DoWithSettings(ctx context.Context, opts *sql.TxOptions, fn func(context.Context) error) error {
	_, transactoinRunning := getFromContext(ctx)
	if transactoinRunning {
		return ErrTransactionInProgress
	}

	tx, err := tm.db.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			err := tx.Rollback()
			if err != nil {
				panic([]any{p, err})
			}
			panic(p) // re-panic after rollback
		}
	}()

	var dbtx DBTX = tx

	for _, opt := range tm.opts {
		dbtx = opt(dbtx)
	}

	txCtx := putInContext(ctx, dbtx)

	if err := fn(txCtx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("transaction error: %w, rollback error: %w", err, rollbackErr)
		}

		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
