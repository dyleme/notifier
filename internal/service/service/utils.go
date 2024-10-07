package service

import (
	"context"
	"errors"

	"github.com/Dyleme/Notifier/pkg/log"
	"github.com/Dyleme/Notifier/pkg/serverrors"
)

func logError(ctx context.Context, err error) {
	var businessError serverrors.BusinessError
	if !errors.As(err, &businessError) {
		log.Ctx(ctx).Error("server error", log.Err(err))
	}
}
