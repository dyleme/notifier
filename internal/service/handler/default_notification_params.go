package handler

import (
	"net/http"

	"github.com/Dyleme/Notifier/internal/authorization/authmiddleware"
	"github.com/Dyleme/Notifier/internal/service/handler/api"
	"github.com/Dyleme/Notifier/pkg/http/requests"
	"github.com/Dyleme/Notifier/pkg/http/responses"
)

func (t TaskHandler) GetDefaultNotificationParams(w http.ResponseWriter, r *http.Request) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	notifParams, err := t.serv.GetDefaultNotificationParams(r.Context(), userID)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	apiNotifParams := mapAPINotificationParams(notifParams)

	responses.JSON(w, http.StatusOK, apiNotifParams)
}

func (t TaskHandler) UpdateDefaultNotificationParams(w http.ResponseWriter, r *http.Request) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)
	}

	var body api.UpdateDefaultNotificationParamsJSONRequestBody
	err = requests.Bind(r, &body)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	notifParams, err := mapDomainNotificationParams(body)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	responses.JSON(w, http.StatusOK, mapAPINotificationParams(np))
}
