package service

import (
	"context"

	"github.com/Dyleme/Notifier/internal/lib/log"
	"github.com/Dyleme/Notifier/internal/lib/serverrors"
)

func logError(ctx context.Context, err error) {
	if _, ok := err.(serverrors.BusinessError); !ok {
		log.Ctx(ctx).Error("server error", log.Err(err))

	}
}
