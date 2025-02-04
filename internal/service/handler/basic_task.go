package handler

import (
	"net/http"

	"github.com/Dyleme/Notifier/internal/authorization/authmiddleware"
	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/internal/service/handler/api"
	"github.com/Dyleme/Notifier/internal/service/service"
	"github.com/Dyleme/Notifier/pkg/http/requests"
	"github.com/Dyleme/Notifier/pkg/http/responses"
	"github.com/Dyleme/Notifier/pkg/utils"
)

func (t TaskHandler) ListBasicTasks(w http.ResponseWriter, r *http.Request, params api.ListBasicTasksParams) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	basicTasks, err := t.serv.ListBasicTasks(r.Context(), userID, service.ListFilterParams{
		ListParams: parseListParams(params.Offset, params.Limit),
		TagIDs:     utils.ZeroIfNil(params.TagIDs),
	})
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	apiBasicTasks := utils.DtoSlice(basicTasks, mapAPIBasicTask)

	responses.JSON(w, http.StatusOK, apiBasicTasks)
}

func (t TaskHandler) CreateBasicTask(w http.ResponseWriter, r *http.Request) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	var body api.CreateBasicTaskJSONRequestBody
	err = requests.Bind(r, &body)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	notifParams, err := mapDomainNotificationParams(body.NotificationParams)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	basicTask := domains.BasicTask{
		ID:                 0,
		UserID:             userID,
		Text:               body.Text,
		Description:        body.Description,
		Start:              body.SendTime,
		NotificationParams: notifParams,
		Tags:               mapDomainTags(body.Tags, userID),
		Notify:             body.Notify,
	}

	createdTask, err := t.serv.CreateBasicTask(r.Context(), basicTask)
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	apiBasicTask := mapAPIBasicTask(createdTask)

	responses.JSON(w, http.StatusCreated, apiBasicTask)
}

func (t TaskHandler) GetBasicTask(w http.ResponseWriter, r *http.Request, taskID int) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	basicTask, err := t.serv.GetBasicTask(r.Context(), userID, taskID)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	apiBasicTask := mapAPIBasicTask(basicTask)

	responses.JSON(w, http.StatusOK, apiBasicTask)
}

func (t TaskHandler) UpdateBasicTask(w http.ResponseWriter, r *http.Request, taskID int) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	var body api.UpdateBasicTaskJSONRequestBody
	err = requests.Bind(r, &body)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	notifParams, err := mapDomainNotificationParams(body.NotificationParams)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	basicTask, err := t.serv.UpdateBasicTask(r.Context(), domains.BasicTask{
		ID:                 taskID,
		UserID:             userID,
		Text:               body.Text,
		Description:        body.Description,
		Start:              body.SendTime,
		NotificationParams: notifParams,
		Tags:               mapDomainTags(body.Tags, userID),
		Notify:             body.Notify,
	}, userID)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	apiBasicTask := mapAPIBasicTask(basicTask)

	responses.JSON(w, http.StatusOK, apiBasicTask)
}

func mapAPIBasicTask(basicTask domains.BasicTask) api.BasicTask {
	return api.BasicTask{
		Description:        &basicTask.Description,
		Id:                 basicTask.ID,
		NotificationParams: utils.Ptr(mapAPINotificationParams(basicTask.NotificationParams)),
		SendTime:           basicTask.Start,
		Text:               basicTask.Text,
		Tags:               mapAPITags(basicTask.Tags),
		Notify:             basicTask.Notify,
	}
}
