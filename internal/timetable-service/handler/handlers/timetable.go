package handlers

import (
	"net/http"
	"time"

	"github.com/Dyleme/Notifier/internal/authorization/authmiddleware"
	"github.com/Dyleme/Notifier/internal/lib/http/requests"
	"github.com/Dyleme/Notifier/internal/lib/http/responses"
	"github.com/Dyleme/Notifier/internal/timetable-service/handler/timetableapi"
	"github.com/Dyleme/Notifier/internal/timetable-service/models"
	"github.com/Dyleme/Notifier/internal/timetable-service/service"
)

func (t TimetableHandler) ListTimetableTasks(w http.ResponseWriter, r *http.Request, params timetableapi.ListTimetableTasksParams) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)
		return
	}
	var (
		from, to time.Time
	)

	if params.From == nil {
		from = time.Time{}
	}
	if params.To == nil {
		to = time.Now().Add(30 * 24 * time.Hour)
	}

	tasks, err := t.serv.ListTimetableTasksInPeriod(r.Context(), userID, from, to)
	if err != nil {
		responses.KnownError(w, err)
		return
	}

	responses.JSON(w, http.StatusOK, tasks)
}

func (t TimetableHandler) PostTimetableSetTaskID(w http.ResponseWriter, r *http.Request, taskID int) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)
		return
	}

	var setTimetableTask timetableapi.SetTimetableTaskReqBody
	err = requests.Bind(r, &setTimetableTask)
	if err != nil {
		responses.Error(w, http.StatusBadRequest, err)
		return
	}
	description := ""
	if setTimetableTask.Description != nil {
		description = *setTimetableTask.Description
	}

	tt, err := t.serv.AddTaskToTimetable(r.Context(), userID, taskID, setTimetableTask.Start, description)
	if err != nil {
		responses.KnownError(w, err)
		return
	}

	apiTT := mapAPITimetableTask(tt)

	responses.JSON(w, http.StatusOK, apiTT)
}

func mapAPITimetableTask(tt models.TimetableTask) timetableapi.TimetableTask {
	return timetableapi.TimetableTask{
		Description: &tt.Description,
		Done:        tt.Done,
		Finish:      tt.Finish,
		Id:          tt.ID,
		Start:       tt.Start,
		TaskId:      tt.TaskID,
		Text:        tt.Text,
	}
}

func (t TimetableHandler) GetTimetableTask(w http.ResponseWriter, r *http.Request, timetableTaskID int) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)
		return
	}

	tt, err := t.serv.GetTimetableTask(r.Context(), userID, timetableTaskID)
	if err != nil {
		responses.KnownError(w, err)
		return
	}

	apiTT := mapAPITimetableTask(tt)

	responses.JSON(w, http.StatusOK, apiTT)
}

func (t TimetableHandler) UpdateTimetableTask(w http.ResponseWriter, r *http.Request, timetableTaskID int) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)
		return
	}

	var updateBody timetableapi.UpdateTimetableTaskJSONRequestBody
	err = requests.Bind(r, &updateBody)
	if err != nil {
		responses.KnownError(w, err)
		return
	}
	description := ""
	if updateBody.Description != nil {
		description = *updateBody.Description
	}

	tt, err := t.serv.UpdateTimetable(r.Context(), service.UpdateTimetableParams{
		ID:          timetableTaskID,
		UserID:      userID,
		Description: description,
		Start:       updateBody.Start,
		Done:        updateBody.Done,
	})
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)
		return
	}

	apiTT := mapAPITimetableTask(tt)

	responses.JSON(w, http.StatusOK, apiTT)
}

func mapCreateTimetableTask(body timetableapi.CreateTimetableTaskReqBody, userID int) (models.Task, models.TimetableTask) {
	description := ""
	if body.Description != nil {
		description = *body.Description
	}
	requiredTime := time.Duration(body.RequiredTime * int(time.Minute))
	task := models.Task{ //nolint:exhaustruct //creation object we don't know ids
		UserID:       userID,
		Text:         body.Message,
		RequiredTime: requiredTime,
		Periodic:     body.Periodic,
	}

	timetableTask := models.TimetableTask{ //nolint:exhaustruct //creation object we don't know ids
		UserID:      userID,
		Text:        body.Message,
		Description: description,
		Start:       body.Start,
		Finish:      body.Start.Add(requiredTime),
		Done:        false,
	}

	return task, timetableTask
}

func (t TimetableHandler) CreateTimetableTask(w http.ResponseWriter, r *http.Request) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)
		return
	}

	var createBody timetableapi.CreateTimetableTaskReqBody
	err = requests.Bind(r, &createBody)
	if err != nil {
		responses.KnownError(w, err)
		return
	}
	task, timetableTask := mapCreateTimetableTask(createBody, userID)
	tt, err := t.serv.CreateTimetableTask(r.Context(), task, timetableTask)
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)
		return
	}

	apiTT := mapAPITimetableTask(tt)

	responses.JSON(w, http.StatusOK, apiTT)
}
