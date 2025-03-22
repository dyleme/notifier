package mock

import (
	"context"

	"github.com/avito-tech/go-transaction-manager/trm"
)

type MockTXManager struct{}

func (tm MockTXManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return tm.DoWithSettings(ctx, nil, fn)
}

func (tm MockTXManager) DoWithSettings(ctx context.Context, _ trm.Settings, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
