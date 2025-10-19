package txmanager

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// TxManager manages database transactions.
type TxManager struct {
	db *sql.DB
}

// txKey is a custom type for context key to avoid collisions.
type txKey struct{}

var (
	// ErrNoTransaction is returned when no transaction is found in context.
	ErrNoTransaction = errors.New("no transaction in context")
	// ErrTransactionInProgress is returned when trying to start a transaction while one is already active.
	ErrTransactionInProgress = errors.New("transaction already in progress")
)

// New creates a new transaction manager.
func New(db *sql.DB) *TxManager {
	return &TxManager{db: db}
}

// getTxFromContext retrieves transaction from context.
func getTxFromContext(ctx context.Context) *sql.Tx {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}

	return nil
}

// DBTX interface that both *sql.DB and *sql.Tx implement.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

func (tm *TxManager) Do(ctx context.Context, fn func(context.Context) error) error {
	return tm.DoWithSettings(ctx, nil, fn)
}

// WithTransaction executes a function within a transaction.
func (tm *TxManager) DoWithSettings(ctx context.Context, opts *sql.TxOptions, fn func(context.Context) error) error {
	tx := getTxFromContext(ctx)
	if tx != nil {
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

	txCtx := context.WithValue(ctx, txKey{}, tx)

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

type Getter struct {
	db *sql.DB
}

func NewGetter(db *sql.DB) *Getter {
	return &Getter{
		db: db,
	}
}

func (tg *Getter) GetTx(ctx context.Context) DBTX {
	tx := getTxFromContext(ctx)
	if tx != nil {
		return tx
	}

	return tg.db
}
