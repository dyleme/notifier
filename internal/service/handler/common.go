package handler

import (
	"time"

	domain "github.com/Dyleme/Notifier/internal/domain"
	serverrors "github.com/Dyleme/Notifier/internal/domain/apperr"
	"github.com/Dyleme/Notifier/internal/service/handler/api"
	"github.com/Dyleme/Notifier/internal/service/service"
	"github.com/Dyleme/Notifier/pkg/model"
	"github.com/Dyleme/Notifier/pkg/utils/ptr"
	"github.com/Dyleme/Notifier/pkg/utils/slice"
)

const timeDay = 24 * time.Hour

func parseListParams(offsetParam *api.OffsetParam, limitParam *api.LimitParam) service.ListParams {
	offset := 0
	limit := 10

	if offsetParam != nil {
		offset = int(*offsetParam)
	}

	if limitParam != nil {
		limit = int(*limitParam)
	}

	return service.ListParams{
		Offset: offset,
		Limit:  limit,
	}
}

func parseTimeParams(from, to *time.Time) model.TimeBorders {
	if from == nil && to == nil {
		return model.NewInfinite()
	}
	if to != nil && from == nil {
		return model.NewInfiniteLower(*to)
	}
	if from != nil && to == nil {
		return model.NewInfiniteUpper(*from)
	}

	return model.New(*from, *to)
}

func mapAPINotificationParams(params domain.NotificationParams) api.NotificationParams {
	return api.NotificationParams{
		NotificationChannel: api.NotificationChannel{
			Cmd:      ptr.NilIfZero(params.Params.Cmd),
			Telegram: ptr.NilIfZero(params.Params.Telegram),
			Webhook:  ptr.NilIfZero(params.Params.Webhook),
		},
		Period: params.Period.String(),
	}
}

func mapDomainNotificationParams(np *api.NotificationParams) (domain.NotificationParams, error) {
	if np == nil {
		return domain.NotificationParams{}, nil
	}
	period, err := time.ParseDuration(np.Period)
	if err != nil {
		return domain.NotificationParams{}, serverrors.NewMappingError(err, "notificationParams.period") //nolint:wrapcheck // standart package error
	}

	return domain.NotificationParams{
		Period: period,
		Params: domain.Params{
			Telegram: ptr.ZeroIfNil(np.NotificationChannel.Telegram),
			Webhook:  ptr.ZeroIfNil(np.NotificationChannel.Webhook),
			Cmd:      ptr.ZeroIfNil(np.NotificationChannel.Cmd),
		},
	}, nil
}

func mapDomainTags(ts []api.Tag, userID int) []domain.Tag {
	return slice.DtoSlice(ts, func(t api.Tag) domain.Tag {
		return domain.Tag{
			ID:     t.Id,
			UserID: userID,
			Name:   t.Name,
		}
	})
}

func mapAPITags(ts []domain.Tag) []api.Tag {
	return slice.DtoSlice(ts, func(t domain.Tag) api.Tag {
		return api.Tag{
			Id:   t.ID,
			Name: t.Name,
		}
	})
}
