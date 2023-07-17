package sqldatabase

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPGX(ctx context.Context, connectionString string) (*pgxpool.Pool, error) {
	conn, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		return nil, err
	}

	err = conn.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
