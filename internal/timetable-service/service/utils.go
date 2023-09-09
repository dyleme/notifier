package service

import (
	"context"
	"errors"

	"github.com/Dyleme/Notifier/internal/lib/log"
	"github.com/Dyleme/Notifier/internal/lib/serverrors"
)

func logError(ctx context.Context, err error) {
	var businessError serverrors.BusinessError
	if !errors.As(err, &businessError) {
		log.Ctx(ctx).Error("server error", log.Err(err))
	}
}
