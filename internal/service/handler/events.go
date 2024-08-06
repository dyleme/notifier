package handler

import (
	"fmt"
	"net/http"

	"github.com/Dyleme/Notifier/internal/authorization/authmiddleware"
	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/internal/service/handler/api"
	"github.com/Dyleme/Notifier/internal/service/service"
	"github.com/Dyleme/Notifier/pkg/http/requests"
	"github.com/Dyleme/Notifier/pkg/http/responses"
	"github.com/Dyleme/Notifier/pkg/log"
	"github.com/Dyleme/Notifier/pkg/serverrors"
	"github.com/Dyleme/Notifier/pkg/utils"
)

func (t TaskHandler) ListEvents(w http.ResponseWriter, r *http.Request, params api.ListEventsParams) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	events, err := t.serv.ListEvents(r.Context(), userID, service.ListEventsFilterParams{
		TimeBorders: parseTimeParams(params.From, params.To),
		ListParams:  parseListParams(params.Offset, params.Limit),
		Tags:        utils.ZeroIfNil(params.TagIDs),
	})
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	apiEvents, err := utils.DtoErrorSlice(events, mapAPIEvent)
	if err != nil {
		responses.KnownError(w, err)
		log.Ctx(r.Context()).Error("map events error", log.Err(err))

		return
	}

	responses.JSON(w, http.StatusOK, apiEvents)
}

func (t TaskHandler) PostTaskSetTaskID(w http.ResponseWriter, _ *http.Request, taskID int) {
	responses.JSON(w, http.StatusOK, taskID)
}

func mapAPIEvent(event domains.Event) (api.Event, error) {
	apiTaskType, err := mapAPITaskType(event.TaskType)
	if err != nil {
		return api.Event{}, err
	}
	apiNotificationParams := mapAPINotificationParams(event.NotificationParams)

	return api.Event{
		Description:        &event.Description,
		Done:               event.Done,
		FirstSendTime:      event.FirstSendTime,
		Id:                 event.ID,
		NextSendTime:       event.NextSendTime,
		NotificationParams: apiNotificationParams,
		TaskID:             event.TaskID,
		TaskType:           apiTaskType,
		Text:               event.Text,
		Tags:               mapAPITags(event.Tags),
	}, nil
}

func mapAPITaskType(taskType domains.TaskType) (api.TaskType, error) {
	switch taskType {
	case domains.PeriodicTaskType:
		return api.Periodic, nil
	case domains.BasicTaskType:
		return api.Basic, nil
	default:
		return "", serverrors.NewServiceError(fmt.Errorf("unknown task type: %s", taskType))
	}
}

func (t TaskHandler) GetEvent(w http.ResponseWriter, r *http.Request, eventID int) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	event, err := t.serv.GetEvent(r.Context(), eventID, userID)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	apiTask, err := mapAPIEvent(event)
	if err != nil {
		log.Ctx(r.Context()).Error("map events error", log.Err(err))
		responses.KnownError(w, err)

		return
	}

	responses.JSON(w, http.StatusOK, apiTask)
}

func (t TaskHandler) RescheduleEvent(w http.ResponseWriter, r *http.Request, eventID int) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	var updateBody api.RescheduleEventJSONRequestBody
	err = requests.Bind(r, &updateBody)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	err = t.serv.ReschedulEventToTime(r.Context(), eventID, userID, updateBody.NextSendTime)
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	responses.Status(w, http.StatusOK)
}

func (t TaskHandler) SetEventDoneStatus(w http.ResponseWriter, r *http.Request, eventID int) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	var updateBody api.SetEventDoneStatusJSONRequestBody
	err = requests.Bind(r, &updateBody)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	err = t.serv.SetEventDoneStatus(r.Context(), eventID, userID, updateBody.Done)
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	responses.Status(w, http.StatusOK)
}
