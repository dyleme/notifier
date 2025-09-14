package sqldatabase

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func NewSQLite(ctx context.Context, filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	err = db.PingContext(ctx)
	if err != nil {
		fmt.Errorf("ping: %w", err)
	}

	return db, nil
}
