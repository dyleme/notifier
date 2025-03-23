package handler

import (
	"net/http"
	"time"

	"github.com/Dyleme/Notifier/internal/authorization/authmiddleware"
	"github.com/Dyleme/Notifier/internal/domain"
	serverrors "github.com/Dyleme/Notifier/internal/domain/apperr"
	"github.com/Dyleme/Notifier/internal/service/handler/api"
	"github.com/Dyleme/Notifier/internal/service/handler/request"
	"github.com/Dyleme/Notifier/internal/service/handler/response"
	"github.com/Dyleme/Notifier/internal/service/service"
	"github.com/Dyleme/Notifier/pkg/utils/ptr"
	"github.com/Dyleme/Notifier/pkg/utils/slice"
)

func (t TaskHandler) ListPeriodicTasks(w http.ResponseWriter, r *http.Request, params api.ListPeriodicTasksParams) {
	ctx := r.Context()
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	listParams := parseListParams(params.Offset, params.Limit)

	periodicTasks, err := t.serv.ListPeriodicTasks(r.Context(), userID, service.ListFilterParams{
		ListParams: listParams,
		TagIDs:     ptr.ZeroIfNil(params.TagIDs),
	})
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	apiPeriodicTasks := slice.Dto(periodicTasks, mapAPIPeriodicTask)

	response.JSON(ctx, w, http.StatusOK, apiPeriodicTasks)
}

func (t TaskHandler) CreatePeriodicTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	var body api.CreatePeriodicTaskJSONRequestBody
	err = request.Bind(r, &body)
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	notifParams, err := mapDomainNotificationParams(body.NotificationParams)
	if err != nil {
		response.Error(ctx, w, err)

		return
	}
	start, err := time.ParseDuration(body.Start)
	if err != nil {
		response.Error(ctx, w, serverrors.ParsingError{
			Cause: err,
			Field: "start",
		})

		return
	}

	basicTask := domain.PeriodicTask{
		ID:                 0,
		Text:               body.Text,
		Description:        ptr.ZeroIfNil(body.Description),
		UserID:             userID,
		Start:              start,
		SmallestPeriod:     24 * time.Hour * time.Duration(body.SmallestPeriod),
		BiggestPeriod:      24 * time.Hour * time.Duration(body.BiggestPeriod),
		NotificationParams: notifParams,
		Tags:               mapDomainTags(body.Tags, userID),
		Notify:             body.Notify,
	}

	createdTask, err := t.serv.CreatePeriodicTask(r.Context(), basicTask, userID)
	if err != nil {
		response.Error(ctx, w, err)

		return
	}
	apiPeriodicTask := mapAPIPeriodicTask(createdTask)

	response.JSON(ctx, w, http.StatusCreated, apiPeriodicTask)
}

func (t TaskHandler) GetPeriodicTask(w http.ResponseWriter, r *http.Request, taskID int) {
	ctx := r.Context()
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	periodicTask, err := t.serv.GetPeriodicTask(r.Context(), taskID, userID)
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	apiPeriodicTask := mapAPIPeriodicTask(periodicTask)

	response.JSON(ctx, w, http.StatusOK, apiPeriodicTask)
}

func (t TaskHandler) UpdatePeriodicTask(w http.ResponseWriter, r *http.Request, taskID int) {
	ctx := r.Context()
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	var body api.UpdatePeriodicTaskJSONRequestBody
	err = request.Bind(r, &body)
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	notifParams, err := mapDomainNotificationParams(body.NotificationParams)
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	start, err := time.ParseDuration(body.Start)
	if err != nil {
		response.Error(ctx, w, serverrors.ParsingError{
			Cause: err,
			Field: "start",
		})

		return
	}

	err = t.serv.UpdatePeriodicTask(r.Context(), domain.PeriodicTask{
		ID:                 taskID,
		Text:               body.Text,
		Description:        ptr.ZeroIfNil(body.Description),
		UserID:             userID,
		Start:              start,
		SmallestPeriod:     time.Duration(body.SmallestPeriod) * 24 * time.Hour,
		BiggestPeriod:      time.Duration(body.BiggestPeriod) * 24 * time.Hour,
		NotificationParams: notifParams,
		Tags:               mapDomainTags(body.Tags, userID),
		Notify:             body.Notify,
	}, userID)
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	response.Status(w, http.StatusOK)
}

func mapAPIPeriodicTask(pt domain.PeriodicTask) api.PeriodicTask {
	return api.PeriodicTask{
		BiggestPeriod:      int(pt.BiggestPeriod / timeDay),
		Description:        &pt.Description,
		Id:                 pt.ID,
		NotificationParams: ptr.On(mapAPINotificationParams(pt.NotificationParams)),
		Notify:             pt.Notify,
		SmallestPeriod:     int(pt.SmallestPeriod / timeDay),
		Start:              pt.Start.String(),
		Tags:               mapAPITags(pt.Tags),
		Text:               pt.Text,
	}
}
