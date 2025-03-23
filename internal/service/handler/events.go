package handler

import (
	"fmt"
	"net/http"

	"github.com/Dyleme/Notifier/internal/authorization/authmiddleware"
	"github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/internal/service/handler/api"
	"github.com/Dyleme/Notifier/internal/service/handler/request"
	"github.com/Dyleme/Notifier/internal/service/handler/response"
	"github.com/Dyleme/Notifier/internal/service/service"
	"github.com/Dyleme/Notifier/pkg/log"
	"github.com/Dyleme/Notifier/pkg/utils/ptr"
	"github.com/Dyleme/Notifier/pkg/utils/slice"
)

func (t TaskHandler) ListEvents(w http.ResponseWriter, r *http.Request, params api.ListEventsParams) {
	ctx := r.Context()
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	events, err := t.serv.ListEvents(r.Context(), userID, service.ListEventsFilterParams{
		TimeBorders: parseTimeParams(params.From, params.To),
		ListParams:  parseListParams(params.Offset, params.Limit),
		Tags:        ptr.ZeroIfNil(params.TagIDs),
	})
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	apiEvents, err := slice.DtoError(events, mapAPIEvent)
	if err != nil {
		response.Error(ctx, w, err)
		log.Ctx(r.Context()).Error("map events error", log.Err(err))

		return
	}

	response.JSON(ctx, w, http.StatusOK, apiEvents)
}

func (t TaskHandler) PostTaskSetTaskID(w http.ResponseWriter, r *http.Request, taskID int) {
	response.JSON(r.Context(), w, http.StatusOK, taskID)
}

func mapAPIEvent(event domain.Event) (api.Event, error) {
	apiTaskType, err := mapAPITaskType(event.TaskType)
	if err != nil {
		return api.Event{}, err
	}
	apiNotificationParams := mapAPINotificationParams(event.NotificationParams)

	return api.Event{
		Description:        &event.Description,
		Done:               event.Done,
		FirstSendTime:      event.FirstSend,
		Id:                 event.ID,
		NextSendTime:       event.NextSend,
		NotificationParams: &apiNotificationParams,
		TaskID:             event.TaskID,
		TaskType:           apiTaskType,
		Text:               event.Text,
		Tags:               mapAPITags(event.Tags),
		Notify:             event.Notify,
	}, nil
}

func mapAPITaskType(taskType domain.TaskType) (api.TaskType, error) {
	switch taskType {
	case domain.PeriodicTaskType:
		return api.Periodic, nil
	case domain.BasicTaskType:
		return api.Basic, nil
	default:
		return "", fmt.Errorf("unknown task type: %s", taskType)
	}
}

func (t TaskHandler) GetEvent(w http.ResponseWriter, r *http.Request, eventID int) {
	ctx := r.Context()
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	event, err := t.serv.GetEvent(r.Context(), eventID, userID)
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	apiTask, err := mapAPIEvent(event)
	if err != nil {
		log.Ctx(r.Context()).Error("map events error", log.Err(err))
		response.Error(ctx, w, err)

		return
	}

	response.JSON(ctx, w, http.StatusOK, apiTask)
}

func (t TaskHandler) RescheduleEvent(w http.ResponseWriter, r *http.Request, eventID int) {
	ctx := r.Context()
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	var updateBody api.RescheduleEventJSONRequestBody
	err = request.Bind(r, &updateBody)
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	err = t.serv.ReschedulEventToTime(r.Context(), eventID, userID, updateBody.NextSendTime)
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	response.Status(w, http.StatusOK)
}

func (t TaskHandler) SetEventDoneStatus(w http.ResponseWriter, r *http.Request, eventID int) {
	ctx := r.Context()
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	var updateBody api.SetEventDoneStatusJSONRequestBody
	err = request.Bind(r, &updateBody)
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	err = t.serv.SetEventDoneStatus(r.Context(), eventID, userID, updateBody.Done)
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	response.Status(w, http.StatusOK)
}
