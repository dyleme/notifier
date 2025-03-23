package handler

import (
	"net/http"

	"github.com/Dyleme/Notifier/internal/authorization/authmiddleware"
	"github.com/Dyleme/Notifier/internal/service/handler/api"
	"github.com/Dyleme/Notifier/internal/service/handler/request"
	"github.com/Dyleme/Notifier/internal/service/handler/response"
)

func (t TaskHandler) GetDefaultNotificationParams(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	notifParams, err := t.serv.GetDefaultNotificationParams(r.Context(), userID)
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	apiNotifParams := mapAPINotificationParams(notifParams)

	response.JSON(ctx, w, http.StatusOK, apiNotifParams)
}

func (t TaskHandler) UpdateDefaultNotificationParams(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		response.Error(ctx, w, err)
	}

	var body api.UpdateDefaultNotificationParamsJSONRequestBody
	err = request.Bind(r, &body)
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	domainNp, err := mapDomainNotificationParams(&body)
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	np, err := t.serv.SetDefaultNotificationParams(r.Context(), domainNp, userID)
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	response.JSON(ctx, w, http.StatusOK, mapAPINotificationParams(np))
}
