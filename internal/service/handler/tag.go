package handler

import (
	"net/http"

	"github.com/Dyleme/Notifier/internal/authorization/authmiddleware"
	"github.com/Dyleme/Notifier/internal/service/handler/api"
	"github.com/Dyleme/Notifier/pkg/http/requests"
	"github.com/Dyleme/Notifier/pkg/http/responses"
)

func (t TaskHandler) ListTags(w http.ResponseWriter, r *http.Request, params api.ListTagsParams) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	tags, err := t.serv.ListTags(r.Context(), userID, parseListParams(params.Offset, params.Limit))
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	responses.JSON(w, http.StatusOK, mapAPITags(tags))
}

func (t TaskHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	var body api.CreateTagJSONBody
	if err := requests.Bind(r, &body); err != nil {
		responses.Error(w, http.StatusBadRequest, err)

		return
	}

	tag, err := t.serv.AddTag(r.Context(), body.Name, userID)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	responses.JSON(w, http.StatusCreated, api.Tag{Id: tag.ID, Name: tag.Name})
}

func (t TaskHandler) GetTag(w http.ResponseWriter, r *http.Request, tagID int) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	tag, err := t.serv.GetTag(r.Context(), tagID, userID)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	responses.JSON(w, http.StatusOK, api.Tag{Id: tag.ID, Name: tag.Name})
}

func (t TaskHandler) UpdateTag(w http.ResponseWriter, r *http.Request, tagID int) {
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err)

		return
	}

	var body api.Tag
	err = requests.Bind(r, &body)
	if err != nil {
		responses.Error(w, http.StatusBadRequest, err)

		return
	}

	err = t.serv.UpdateTag(r.Context(), tagID, body.Name, userID)
	if err != nil {
		responses.KnownError(w, err)

		return
	}

	responses.Status(w, http.StatusOK)
}
