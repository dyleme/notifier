package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dyleme/Notifier/internal/domain/apperr"
	"github.com/dyleme/Notifier/internal/repository/queries/goqueries"
	"github.com/dyleme/Notifier/pkg/database/txmanager"
	"github.com/dyleme/Notifier/pkg/log"
)

type KeyValueRepository struct {
	q      *goqueries.Queries
	getter *txmanager.Getter
}

func NewKeyValueRepository(getter *txmanager.Getter) *KeyValueRepository {
	return &KeyValueRepository{
		q:      goqueries.New(),
		getter: getter,
	}
}

var (
	ErrEmptyValue = errors.New("value is empty")
	ErrEmptyKey   = errors.New("empty key")
)

func (r *KeyValueRepository) PutValue(ctx context.Context, key string, value any) error {
	tx := r.getter.GetTx(ctx)

	if key == "" {
		return ErrEmptyKey
	}

	bts, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if len(bts) == 0 {
		return ErrEmptyValue
	}

	log.Ctx(ctx).Debug("put value", "key", key, "value", value)
	err = r.q.SetValue(ctx, tx, goqueries.SetValueParams{
		Key:   key,
		Value: bts,
	})
	if err != nil {
		return fmt.Errorf("put value: %w", err)
	}

	return nil
}

func (r *KeyValueRepository) GetValue(ctx context.Context, key string, value any) error {
	tx := r.getter.GetTx(ctx)

	bts, err := r.q.GetValue(ctx, tx, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperr.ErrNotFound
		}

		return fmt.Errorf("get value: %w", err)
	}

	err = json.Unmarshal(bts, value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return nil
}

func (r *KeyValueRepository) DeleteValue(ctx context.Context, key string) error {
	tx := r.getter.GetTx(ctx)

	err := r.q.DeleteValue(ctx, tx, key)
	if err != nil {
		return fmt.Errorf("delete value: %w", err)
	}

	return nil
}
