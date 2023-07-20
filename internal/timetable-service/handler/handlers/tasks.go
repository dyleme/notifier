package handlers

import (
	"net/http"
	"time"

	"github.com/Dyleme/Notifier/internal/authorization/authmiddleware"
	"github.com/Dyleme/Notifier/internal/lib/http/requests"
	"github.com/Dyleme/Notifier/internal/lib/http/responses"
	"github.com/Dyleme/Notifier/internal/timetable-service/handler/timetableapi"
	"github.com/Dyleme/Notifier/internal/timetable-service/models"
)

func (t TimetableHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)
		return
	}

	tasks, err := t.serv.GetUserTasks(r.Context(), userID)
	if err != nil {
		responses.KnownError(w, err)
		return
	}
	apiTasks := mapAPITasks(tasks)

	responses.JSON(w, http.StatusOK, apiTasks)
}

func (t TimetableHandler) AddTask(w http.ResponseWriter, r *http.Request) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)
		return
	}

	var addTaskBody timetableapi.AddTaskJSONRequestBody
	err = requests.Bind(r, &addTaskBody)
	if err != nil {
		responses.Error(w, http.StatusBadRequest, err)
		return
	}

	task := mapAddTaskReq(addTaskBody, userID)
	createdTask, err := t.serv.AddTask(r.Context(), task)
	if err != nil {
		responses.KnownError(w, err)
		return
	}

	responses.JSON(w, http.StatusCreated, mapAPITask(createdTask))
}

func mapAddTaskReq(body timetableapi.AddTaskJSONRequestBody, userID int) models.Task {
	return models.Task{ //nolint:exhaustruct // TODO: use separate struct for creation
		UserID:       userID,
		Text:         body.Message,
		RequiredTime: time.Duration(body.RequiredTime * int(time.Minute)),
		Periodic:     body.Periodic,
		Done:         false,
		Archived:     false,
	}
}

func mapUpdateTaskReq(body timetableapi.UpdateTaskReqBody, taskID, userID int) models.Task {
	return models.Task{
		ID:           taskID,
		UserID:       userID,
		Text:         body.Message,
		RequiredTime: time.Duration(body.RequiredTime * int(time.Minute)),
		Periodic:     body.Periodic,
		Done:         body.Done,
		Archived:     body.Archived,
	}
}

func mapAPITask(task models.Task) timetableapi.Task {
	return timetableapi.Task{
		Id:           task.ID,
		Message:      task.Text,
		RequiredTime: int(task.RequiredTime.Minutes()),
		Archived:     task.Archived,
		Done:         task.Done,
		Periodic:     task.Periodic,
	}
}

func mapAPITasks(tasks []models.Task) []timetableapi.Task {
	apiTasks := make([]timetableapi.Task, 0, len(tasks))
	for _, t := range tasks {
		apiTasks = append(apiTasks, mapAPITask(t))
	}

	return apiTasks
}

func (t TimetableHandler) GetTask(w http.ResponseWriter, r *http.Request, taskID int) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)
		return
	}

	task, err := t.serv.GetTask(r.Context(), taskID, userID)
	if err != nil {
		responses.KnownError(w, err)
		return
	}

	responses.JSON(w, http.StatusOK, mapAPITask(task))
}

func (t TimetableHandler) UpdateTask(w http.ResponseWriter, r *http.Request, taskID int) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)
		return
	}

	var reqBody timetableapi.UpdateTaskJSONRequestBody
	err = requests.Bind(r, &reqBody)
	if err != nil {
		responses.Error(w, http.StatusBadRequest, err)
		return
	}

	task := mapUpdateTaskReq(reqBody, userID, taskID)

	err = t.serv.UpdateTask(r.Context(), task)
	if err != nil {
		responses.KnownError(w, err)
		return
	}

	responses.Status(w, http.StatusOK)
}
