package handlers

import (
	"net/http"
	"time"

	"github.com/Dyleme/Notifier/internal/authorization/authmiddleware"
	"github.com/Dyleme/Notifier/internal/lib/http/requests"
	"github.com/Dyleme/Notifier/internal/lib/http/responses"
	"github.com/Dyleme/Notifier/internal/lib/utils/ptr"
	"github.com/Dyleme/Notifier/internal/timetable-service/handler/timetableapi"
	"github.com/Dyleme/Notifier/internal/timetable-service/models"
)

func (t TimetableHandler) GetDefaultNotificationParams(w http.ResponseWriter, r *http.Request) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)
		return
	}

	defParams, err := t.serv.GetDefaultNotificationParams(r.Context(), userID)
	if err != nil {
		responses.KnownError(w, err)
		return
	}

	responses.JSON(w, http.StatusOK, mapNotificationParamsResp(defParams))
}

func (t TimetableHandler) SetDefaultNotificationParams(w http.ResponseWriter, r *http.Request) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)
		return
	}

	var reqParamsBody timetableapi.SetDefaultNotificationParamsJSONRequestBody
	err = requests.Bind(r, &reqParamsBody)
	if err != nil {
		responses.KnownError(w, err)
		return
	}

	params := mapNotificationParams(reqParamsBody)

	defParams, err := t.serv.SetDefaultNotificationParams(r.Context(), params, userID)
	if err != nil {
		responses.KnownError(w, err)
		return
	}

	responses.JSON(w, http.StatusOK, mapNotificationParamsResp(defParams))
}

func mapNotificationParamsResp(p models.NotificationParams) timetableapi.NotificationParams {
	return timetableapi.NotificationParams{
		Info: timetableapi.NotificationInfo{
			Cmd:      ptr.Ptr(true),
			Telegram: ptr.Ptr(true),
			Webhook:  &p.Params.Webhook,
		},
		Period: int(p.Period.Minutes()),
	}
}

func mapNotificationParams(req timetableapi.NotificationParams) models.NotificationParams {
	var (
		webhook string
	)
	if req.Info.Webhook != nil {
		webhook = *req.Info.Webhook
	}
	params := models.NotificationParams{
		Period: time.Duration(req.Period) * time.Minute,
		Params: models.Params{
			Telegram: "",
			Webhook:  webhook,
			Cmd:      "",
		},
	}
	return params
}

func (t TimetableHandler) GetTimetableTaskNotificationParams(w http.ResponseWriter, r *http.Request, timetableTaskID int) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)
		return
	}

	params, err := t.serv.GetNotificationParams(r.Context(), timetableTaskID, userID)
	if err != nil {
		responses.KnownError(w, err)
		return
	}

	responses.JSON(w, http.StatusOK, mapNotificationParamsResp(*params))
}

func (t TimetableHandler) SetTimetableTaskNotificationParams(w http.ResponseWriter, r *http.Request, timetableTaskID int) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)
		return
	}

	var reqParamsBody timetableapi.SetTimetableTaskNotificationParamsJSONRequestBody
	err = requests.Bind(r, &reqParamsBody)
	if err != nil {
		responses.KnownError(w, err)
		return
	}

	params := mapNotificationParams(reqParamsBody)

	res, err := t.serv.SetNotificationParams(r.Context(), timetableTaskID, params, userID)
	if err != nil {
		responses.KnownError(w, err)
		return
	}

	responses.JSON(w, http.StatusOK, mapNotificationParamsResp(res))
}
