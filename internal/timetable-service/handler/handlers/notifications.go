package handlers

import (
	"net/http"
	"time"

	"github.com/Dyleme/Notifier/internal/authorization/authmiddleware"
	"github.com/Dyleme/Notifier/internal/lib/http/requests"
	"github.com/Dyleme/Notifier/internal/lib/http/responses"
	"github.com/Dyleme/Notifier/internal/lib/utils/ptr"
	"github.com/Dyleme/Notifier/internal/timetable-service/domains"
	"github.com/Dyleme/Notifier/internal/timetable-service/handler/timetableapi"
)

func (t EventHandler) GetDefaultNotificationParams(w http.ResponseWriter, r *http.Request) {
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

func (t EventHandler) SetDefaultNotificationParams(w http.ResponseWriter, r *http.Request) {
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

func mapNotificationParamsResp(p domains.NotificationParams) timetableapi.NotificationParams {
	return timetableapi.NotificationParams{
		Info: timetableapi.NotificationInfo{
			Cmd:      ptr.Ptr(true),
			Telegram: &p.Params.Telegram,
			Webhook:  &p.Params.Webhook,
		},
		Period:      int(p.Period.Minutes()),
		DelayedTill: p.DalayedTill,
	}
}

func mapNotificationParams(req timetableapi.NotificationParams) domains.NotificationParams {
	var (
		webhook  string
		telegram int
	)
	if req.Info.Webhook != nil {
		webhook = *req.Info.Webhook
	}
	if req.Info.Telegram != nil {
		telegram = *req.Info.Telegram
	}
	params := domains.NotificationParams{
		Period: time.Duration(req.Period) * time.Minute,
		Params: domains.Params{
			Telegram: telegram,
			Webhook:  webhook,
			Cmd:      "",
		},
		DalayedTill: req.DelayedTill,
	}

	return params
}

func (t EventHandler) GetEventNotificationParams(w http.ResponseWriter, r *http.Request, eventID int) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	params, err := t.serv.GetNotificationParams(r.Context(), eventID, userID)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	responses.JSON(w, http.StatusOK, mapNotificationParamsResp(*params))
}

func (t EventHandler) SetEventNotificationParams(w http.ResponseWriter, r *http.Request, eventID int) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	var reqParamsBody timetableapi.SetEventNotificationParamsJSONRequestBody
	err = requests.Bind(r, &reqParamsBody)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	params := mapNotificationParams(reqParamsBody)

	res, err := t.serv.SetNotificationParams(r.Context(), eventID, params, userID)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	responses.JSON(w, http.StatusOK, mapNotificationParamsResp(res))
}
