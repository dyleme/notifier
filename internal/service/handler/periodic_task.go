package handler

import (
	"net/http"
	"time"

	"github.com/Dyleme/Notifier/internal/authorization/authmiddleware"
	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/internal/service/handler/api"
	"github.com/Dyleme/Notifier/internal/service/service"
	"github.com/Dyleme/Notifier/pkg/http/requests"
	"github.com/Dyleme/Notifier/pkg/http/responses"
	"github.com/Dyleme/Notifier/pkg/serverrors"
	"github.com/Dyleme/Notifier/pkg/utils"
)

func (t TaskHandler) ListPeriodicTasks(w http.ResponseWriter, r *http.Request, params api.ListPeriodicTasksParams) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	listParams := parseListParams(params.Offset, params.Limit)

	periodicTasks, err := t.serv.ListPeriodicTasks(r.Context(), userID, service.ListFilterParams{
		ListParams: listParams,
		TagIDs:     utils.ZeroIfNil(params.TagIDs),
	})
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	apiPeriodicTasks := utils.DtoSlice(periodicTasks, mapAPIPeriodicTask)

	responses.JSON(w, http.StatusOK, apiPeriodicTasks)
}

func (t TaskHandler) CreatePeriodicTask(w http.ResponseWriter, r *http.Request) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	var body api.CreatePeriodicTaskJSONRequestBody
	err = requests.Bind(r, &body)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	notifParams, err := mapPtrDomainNotificationParams(body.NotificationParams)
	if err != nil {
		responses.KnownError(w, err)

		return
	}
	start, err := time.ParseDuration(body.Start)
	if err != nil {
		responses.KnownError(w, serverrors.NewMappingError(err, "start"))

		return
	}

	basicTask := domains.PeriodicTask{
		ID:                 0,
		Text:               body.Text,
		Description:        utils.ZeroIfNil(body.Description),
		UserID:             userID,
		Start:              start,
		SmallestPeriod:     24 * time.Hour * time.Duration(body.SmallestPeriod),
		BiggestPeriod:      24 * time.Hour * time.Duration(body.BiggestPeriod),
		NotificationParams: notifParams,
		Tags:               mapDomainTags(body.Tags, userID),
	}

	createdTask, err := t.serv.CreatePeriodicTask(r.Context(), basicTask, userID)
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}
	apiPeriodicTask := mapAPIPeriodicTask(createdTask)

	responses.JSON(w, http.StatusCreated, apiPeriodicTask)
}

func (t TaskHandler) GetPeriodicTask(w http.ResponseWriter, r *http.Request, taskID int) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	periodicTask, err := t.serv.GetPeriodicTask(r.Context(), taskID, userID)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	apiPeriodicTask := mapAPIPeriodicTask(periodicTask)

	responses.JSON(w, http.StatusOK, apiPeriodicTask)
}

func (t TaskHandler) UpdatePeriodicTask(w http.ResponseWriter, r *http.Request, taskID int) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	var body api.UpdatePeriodicTaskJSONRequestBody
	err = requests.Bind(r, &body)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	notifParams, err := mapPtrDomainNotificationParams(body.NotificationParams)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	start, err := time.ParseDuration(body.Start)
	if err != nil {
		responses.KnownError(w, serverrors.NewMappingError(err, "start"))

		return
	}

	err = t.serv.UpdatePeriodicTask(r.Context(), domains.PeriodicTask{
		ID:                 taskID,
		Text:               body.Text,
		Description:        utils.ZeroIfNil(body.Description),
		UserID:             userID,
		Start:              start,
		SmallestPeriod:     time.Duration(body.SmallestPeriod) * 24 * time.Hour,
		BiggestPeriod:      time.Duration(body.BiggestPeriod) * 24 * time.Hour,
		NotificationParams: notifParams,
		Tags:               mapDomainTags(body.Tags, userID),
	}, userID)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	responses.Status(w, http.StatusOK)
}

func mapAPIPeriodicTask(pt domains.PeriodicTask) api.PeriodicTask {
	return api.PeriodicTask{
		Id:                 pt.ID,
		Description:        &pt.Description,
		NotificationParams: mapPtrAPINotificationParams(pt.NotificationParams),
		BiggestPeriod:      int(pt.BiggestPeriod / timeDay),
		SmallestPeriod:     int(pt.SmallestPeriod / timeDay),
		Start:              pt.Start.String(),
		Text:               pt.Text,
		Tags:               mapAPITags(pt.Tags),
	}
}
