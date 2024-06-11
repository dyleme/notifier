package handler

import (
	"net/http"
	"time"

	"github.com/Dyleme/Notifier/internal/authorization/authmiddleware"
	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/internal/service/handler/api"
	"github.com/Dyleme/Notifier/pkg/http/requests"
	"github.com/Dyleme/Notifier/pkg/http/responses"
	"github.com/Dyleme/Notifier/pkg/utils"
)

func (t TaskHandler) GetDefaultEventParams(w http.ResponseWriter, r *http.Request) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	defParams, err := t.serv.GetDefaultEventParams(r.Context(), userID)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	responses.JSON(w, http.StatusOK, mapEventParamsResp(defParams))
}

func (t TaskHandler) SetDefaultEventParams(w http.ResponseWriter, r *http.Request) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	var reqParamsBody api.SetDefaultEventParamsJSONRequestBody
	err = requests.Bind(r, &reqParamsBody)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	params := mapEventParams(reqParamsBody)

	defParams, err := t.serv.SetDefaultEventParams(r.Context(), params, userID)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	responses.JSON(w, http.StatusOK, mapEventParamsResp(defParams))
}

func mapEventParamsResp(p domains.NotificationParams) api.EventParams {
	return api.EventParams{
		Info: api.EventInfo{
			Cmd:      utils.Ptr(true),
			Telegram: &p.Params.Telegram,
			Webhook:  &p.Params.Webhook,
		},
		Period:      int(p.Period.Minutes()),
		DelayedTill: nil,
	}
}

func mapEventParams(req api.EventParams) domains.NotificationParams {
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
	}

	return params
}

func (t TaskHandler) GetTaskEventParams(w http.ResponseWriter, r *http.Request, _ int) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	params, err := t.serv.GetDefaultEventParams(r.Context(), userID)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	responses.JSON(w, http.StatusOK, mapEventParamsResp(params))
}

func (t TaskHandler) SetTaskEventParams(w http.ResponseWriter, r *http.Request, _ int) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	var reqParamsBody api.SetTaskEventParamsJSONRequestBody
	err = requests.Bind(r, &reqParamsBody)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	params := mapEventParams(reqParamsBody)

	res, err := t.serv.SetDefaultEventParams(r.Context(), params, userID)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	responses.JSON(w, http.StatusOK, mapEventParamsResp(res))
}
