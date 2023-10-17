package sqldatabase

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPGX(ctx context.Context, connectionString string) (*pgxpool.Pool, error) {
	op := "NewPGX: %w"
	conn, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	err = conn.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return conn, nil
}
