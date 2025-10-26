package txmanager

import (
	"context"
	"database/sql"
	"errors"
)

type Option func(DBTX) DBTX

// txKey is a custom type for context key to avoid collisions.
type txKey struct{}

var (
	// ErrNoTransaction is returned when no transaction is found in context.
	ErrNoTransaction = errors.New("no transaction in context")
	// ErrTransactionInProgress is returned when trying to start a transaction while one is already active.
	ErrTransactionInProgress = errors.New("transaction already in progress")
)

// New creates a new transaction manager.
func New(db *sql.DB, opts ...Option) (*TxManager, *Getter) {
	m := newTxManager(db, opts)
	g := newGetter(db, opts)

	return m, g
}

// getTxFromContext retrieves transaction from context.
func getTxFromContext(ctx context.Context) *sql.Tx {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}

	return nil
}

func putInContext(ctx context.Context, tx DBTX) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// DBTX interface that both *sql.DB and *sql.Tx implement.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}
