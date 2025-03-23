package handler

import (
	"net/http"

	"github.com/Dyleme/Notifier/internal/authorization/authmiddleware"
	"github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/internal/service/handler/api"
	"github.com/Dyleme/Notifier/internal/service/handler/request"
	"github.com/Dyleme/Notifier/internal/service/handler/response"
	"github.com/Dyleme/Notifier/internal/service/service"
	"github.com/Dyleme/Notifier/pkg/utils/ptr"
	"github.com/Dyleme/Notifier/pkg/utils/slice"
)

func (t TaskHandler) ListBasicTasks(w http.ResponseWriter, r *http.Request, params api.ListBasicTasksParams) {
	ctx := r.Context()
	userID, err := authmiddleware.UserIDFromCtx(ctx)
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	basicTasks, err := t.serv.ListBasicTasks(r.Context(), userID, service.ListFilterParams{
		ListParams: parseListParams(params.Offset, params.Limit),
		TagIDs:     ptr.ZeroIfNil(params.TagIDs),
	})
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	apiBasicTasks := slice.Dto(basicTasks, mapAPIBasicTask)

	response.JSON(ctx, w, http.StatusOK, apiBasicTasks)
}

func (t TaskHandler) CreateBasicTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	var body api.CreateBasicTaskJSONRequestBody
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

	basicTask := domain.BasicTask{
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
		response.Error(ctx, w, err)

		return
	}

	apiBasicTask := mapAPIBasicTask(createdTask)

	response.JSON(ctx, w, http.StatusCreated, apiBasicTask)
}

func (t TaskHandler) GetBasicTask(w http.ResponseWriter, r *http.Request, taskID int) {
	ctx := r.Context()
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	basicTask, err := t.serv.GetBasicTask(r.Context(), userID, taskID)
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	apiBasicTask := mapAPIBasicTask(basicTask)

	response.JSON(ctx, w, http.StatusOK, apiBasicTask)
}

func (t TaskHandler) UpdateBasicTask(w http.ResponseWriter, r *http.Request, taskID int) {
	ctx := r.Context()
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	var body api.UpdateBasicTaskJSONRequestBody
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

	basicTask, err := t.serv.UpdateBasicTask(r.Context(), domain.BasicTask{
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
		response.Error(ctx, w, err)

		return
	}

	apiBasicTask := mapAPIBasicTask(basicTask)

	response.JSON(ctx, w, http.StatusOK, apiBasicTask)
}

func mapAPIBasicTask(basicTask domain.BasicTask) api.BasicTask {
	return api.BasicTask{
		Description:        &basicTask.Description,
		Id:                 basicTask.ID,
		NotificationParams: ptr.On(mapAPINotificationParams(basicTask.NotificationParams)),
		SendTime:           basicTask.Start,
		Text:               basicTask.Text,
		Tags:               mapAPITags(basicTask.Tags),
		Notify:             basicTask.Notify,
	}
}
