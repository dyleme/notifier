package sqldatabase

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite" // sqlite driver
)

func NewSQLite(ctx context.Context, filepath string) (*sql.DB, func() error, error) {
	_, err := os.Stat(filepath)
	if err != nil {
		return nil, nil, fmt.Errorf("stat: %w", err)
	}
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		return nil, nil, fmt.Errorf("open sqlite: %w", err)
	}

	err = db.PingContext(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("ping: %w", err)
	}

	return db, db.Close, nil
}
