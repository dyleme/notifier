package service

import (
	"context"
	"errors"

	"github.com/Dyleme/Notifier/internal/domain/apperr"
	"github.com/Dyleme/Notifier/pkg/log"
)

func logError(ctx context.Context, err error) {
	var businessError apperr.BusinessError
	if !errors.As(err, &businessError) {
		log.Ctx(ctx).Error("server error", log.Err(err))
	}
}
