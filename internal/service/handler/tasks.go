package handler

import (
	"net/http"

	"github.com/Dyleme/Notifier/internal/authorization/authmiddleware"
	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/internal/service/handler/api"
	"github.com/Dyleme/Notifier/pkg/http/requests"
	"github.com/Dyleme/Notifier/pkg/http/responses"
)

func (t TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request, params api.ListTasksParams) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	listParams := parseListParams(params.Offset, params.Limit)

	tasks, err := t.serv.ListTasks(r.Context(), userID, listParams)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	responses.JSON(w, http.StatusOK, tasks)
}

func (t TaskHandler) PostTaskSetTaskID(w http.ResponseWriter, r *http.Request, taskID int) {
	responses.JSON(w, http.StatusOK, taskID)
}

func mapAPITask(task domains.BasicTask) api.Task {
	return api.Task{
		Description: &task.Description,
		Done:        false,
		Id:          task.ID,
		Start:       task.Start,
		Text:        task.Text,
	}
}

func (t TaskHandler) GetTask(w http.ResponseWriter, r *http.Request, taskID int) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	task, err := t.serv.GetTask(r.Context(), userID, taskID)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	apiTask := mapAPITask(task)

	responses.JSON(w, http.StatusOK, apiTask)
}

func (t TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request, taskID int) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	var updateBody api.UpdateTaskJSONRequestBody
	err = requests.Bind(r, &updateBody)
	if err != nil {
		responses.KnownError(w, err)

		return
	}
	description := ""
	if updateBody.Description != nil {
		description = *updateBody.Description
	}

	task, err := t.serv.UpdateBasicTask(r.Context(), domains.BasicTask{
		ID:                 taskID,
		UserID:             userID,
		Text:               description,
		Description:        description,
		Start:              updateBody.Start,
		NotificationParams: nil,
	}, userID)
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	apiTask := mapAPITask(task)

	responses.JSON(w, http.StatusOK, apiTask)
}

func mapCreateTask(body api.CreateTaskReqBody, userID int) domains.BasicTask {
	description := ""
	if body.Description != nil {
		description = *body.Description
	}

	task := domains.BasicTask{ //nolint:exhaustruct //creation object we don't know ids
		UserID:             userID,
		Text:               body.Message,
		Description:        description,
		Start:              body.Start,
		NotificationParams: nil,
	}

	return task
}

func (t TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	var createBody api.CreateTaskReqBody
	err = requests.Bind(r, &createBody)
	if err != nil {
		responses.KnownError(w, err)

		return
	}
	task := mapCreateTask(createBody, userID)
	createdTask, err := t.serv.CreateTask(r.Context(), task)
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	apiTask := mapAPITask(createdTask)

	responses.JSON(w, http.StatusOK, apiTask)
}
