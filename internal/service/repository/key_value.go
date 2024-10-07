package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	trmpgx "github.com/avito-tech/go-transaction-manager/pgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Dyleme/Notifier/internal/service/repository/queries/goqueries"
	"github.com/Dyleme/Notifier/pkg/serverrors"
)

type KeyValueRepository struct {
	q      *goqueries.Queries
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewKeyValueRepository(db *pgxpool.Pool, getter *trmpgx.CtxGetter) *KeyValueRepository {
	return &KeyValueRepository{
		q:      goqueries.New(),
		db:     db,
		getter: getter,
	}
}

var (
	errEmptyValue = errors.New("value is empty")
	errEmptyKey   = errors.New("empty key")
)

func (r *KeyValueRepository) PutValue(ctx context.Context, key string, value any) error {
	tx := r.getter.DefaultTrOrDB(ctx, r.db)

	if key == "" {
		return serverrors.NewServiceError(errEmptyKey)
	}

	bts, err := json.Marshal(value)
	if err != nil {
		return serverrors.NewServiceError(fmt.Errorf("failed to marshal value: %w", err))
	}

	if len(bts) == 0 {
		return serverrors.NewServiceError(errEmptyValue)
	}

	err = r.q.SetValue(ctx, tx, goqueries.SetValueParams{
		Key:   key,
		Value: bts,
	})
	if err != nil {
		return fmt.Errorf("put value: %w", serverrors.NewRepositoryError(err))
	}

	return nil
}

func (r *KeyValueRepository) GetValue(ctx context.Context, key string, value any) error {
	tx := r.getter.DefaultTrOrDB(ctx, r.db)

	bts, err := r.q.GetValue(ctx, tx, key)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return serverrors.NewNotFoundError(err, "value not found")
		}

		return fmt.Errorf("get value: %w", serverrors.NewRepositoryError(err))
	}

	err = json.Unmarshal(bts, value)
	if err != nil {
		return serverrors.NewServiceError(fmt.Errorf("failed to marshal value: %w", err))
	}

	return nil
}

func (r *KeyValueRepository) DeleteValue(ctx context.Context, key string) error {
	tx := r.getter.DefaultTrOrDB(ctx, r.db)

	err := r.q.DeleteValue(ctx, tx, key)
	if err != nil {
		return fmt.Errorf("delete value: %w", serverrors.NewRepositoryError(err))
	}

	return nil
}
