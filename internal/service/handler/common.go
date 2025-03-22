package handler

import (
	"time"

	domain "github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/internal/service/handler/api"
	"github.com/Dyleme/Notifier/internal/service/service"
	"github.com/Dyleme/Notifier/pkg/serverrors"
	"github.com/Dyleme/Notifier/pkg/utils"
	"github.com/Dyleme/Notifier/pkg/utils/timeborders"
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

func parseTimeParams(from, to *time.Time) timeborders.TimeBorders {
	if from == nil && to == nil {
		return timeborders.NewInfinite()
	}
	if to != nil && from == nil {
		return timeborders.NewInfiniteLower(*to)
	}
	if from != nil && to == nil {
		return timeborders.NewInfiniteUpper(*from)
	}

	return timeborders.New(*from, *to)
}

func mapAPINotificationParams(params domain.NotificationParams) api.NotificationParams {
	return api.NotificationParams{
		NotificationChannel: api.NotificationChannel{
			Cmd:      utils.NilIfZero(params.Params.Cmd),
			Telegram: utils.NilIfZero(params.Params.Telegram),
			Webhook:  utils.NilIfZero(params.Params.Webhook),
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
			Telegram: utils.ZeroIfNil(np.NotificationChannel.Telegram),
			Webhook:  utils.ZeroIfNil(np.NotificationChannel.Webhook),
			Cmd:      utils.ZeroIfNil(np.NotificationChannel.Cmd),
		},
	}, nil
}

func mapDomainTags(ts []api.Tag, userID int) []domain.Tag {
	return utils.DtoSlice(ts, func(t api.Tag) domain.Tag {
		return domain.Tag{
			ID:     t.Id,
			UserID: userID,
			Name:   t.Name,
		}
	})
}

func mapAPITags(ts []domain.Tag) []api.Tag {
	return utils.DtoSlice(ts, func(t domain.Tag) api.Tag {
		return api.Tag{
			Id:   t.ID,
			Name: t.Name,
		}
	})
}
