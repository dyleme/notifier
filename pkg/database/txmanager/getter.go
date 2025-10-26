package txmanager

import (
	"context"
	"database/sql"
)

type Getter struct {
	db   DBTX
	opts []Option
}

func newGetter(db *sql.DB, opts []Option) *Getter {
	g := &Getter{
		opts: opts,
		db:   db,
	}

	for _, opt := range opts {
		g.db = opt(g.db)
	}

	return g
}

func (tg *Getter) GetTx(ctx context.Context) DBTX {
	tx := getTxFromContext(ctx)
	if tx != nil {
		return tx
	}

	return tg.db
}
