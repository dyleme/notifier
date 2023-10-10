package handler

import (
	"github.com/Dyleme/Notifier/internal/service/handler/api"
	"github.com/Dyleme/Notifier/internal/service/service"
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
