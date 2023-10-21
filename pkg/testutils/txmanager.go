package testutils

import (
	"context"

	"github.com/avito-tech/go-transaction-manager/trm"
)

func TxManager(ctx context.Context, _ trm.Settings) (context.Context, trm.Transaction, error) {
	return ctx, TransactionMock{}, nil
}

type TransactionMock struct{}

func (t TransactionMock) Transaction() interface{} {
	return nil
}

func (t TransactionMock) Commit(_ context.Context) error {
	return nil
}

func (t TransactionMock) Rollback(_ context.Context) error {
	return nil
}

func (t TransactionMock) IsActive() bool {
	return true
}
