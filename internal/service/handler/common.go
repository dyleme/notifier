package handler

import (
	"time"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/internal/service/handler/api"
	"github.com/Dyleme/Notifier/internal/service/service"
	"github.com/Dyleme/Notifier/pkg/serverrors"
	"github.com/Dyleme/Notifier/pkg/utils"
	"github.com/Dyleme/Notifier/pkg/utils/timeborders"
)

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

func mapAPINotificationParams(params domains.NotificationParams) api.NotificationParams {
	return api.NotificationParams{
		NotificationChannel: api.NotificationChannel{
			Cmd:      utils.NilIfZero(params.Params.Cmd),
			Telegram: utils.NilIfZero(params.Params.Telegram),
			Webhook:  utils.NilIfZero(params.Params.Webhook),
		},
		Period: params.Period.String(),
	}
}

func mapPtrAPINotificationParams(params *domains.NotificationParams) *api.NotificationParams {
	if params == nil {
		return nil
	}

	return utils.Ptr(mapAPINotificationParams(*params))
}

func mapPtrDomainNotificationParams(np *api.NotificationParams) (*domains.NotificationParams, error) {
	if np == nil {
		return nil, nil
	}
	notifParams, err := mapDomainNotificationParams(*np)
	if err != nil {
		return nil, err
	}

	return &notifParams, nil
}

func mapDomainNotificationParams(np api.NotificationParams) (domains.NotificationParams, error) {
	period, err := time.ParseDuration(np.Period)
	if err != nil {
		return domains.NotificationParams{}, serverrors.NewMappingError(err, "notificationParams.period")
	}
	return domains.NotificationParams{
		Period: period,
		Params: domains.Params{
			Telegram: utils.ZeroIfNil(np.NotificationChannel.Telegram),
			Webhook:  utils.ZeroIfNil(np.NotificationChannel.Webhook),
			Cmd:      utils.ZeroIfNil(np.NotificationChannel.Cmd),
		},
	}, nil
}
