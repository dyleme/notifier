package mock

import (
	"context"

	"github.com/avito-tech/go-transaction-manager/trm"
)

type TxManager struct{}

func (tm TxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return tm.DoWithSettings(ctx, nil, fn)
}

func (tm TxManager) DoWithSettings(ctx context.Context, _ trm.Settings, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
