package sqldatabase

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib" // database driver
)

func NewSQL(ctx context.Context, connectionString string) (*sql.DB, error) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, err
	}

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
