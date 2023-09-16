package sqldatabase

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // database driver
)

func NewSQL(ctx context.Context, connectionString string) (*sql.DB, error) {
	op := "NewSQL: %w"
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return db, nil
}
