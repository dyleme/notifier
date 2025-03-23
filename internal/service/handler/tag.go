package handler

import (
	"net/http"

	"github.com/Dyleme/Notifier/internal/authorization/authmiddleware"
	"github.com/Dyleme/Notifier/internal/service/handler/api"
	"github.com/Dyleme/Notifier/internal/service/handler/request"
	"github.com/Dyleme/Notifier/internal/service/handler/response"
)

func (t TaskHandler) ListTags(w http.ResponseWriter, r *http.Request, params api.ListTagsParams) {
	ctx := r.Context()
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	tags, err := t.serv.ListTags(r.Context(), userID, parseListParams(params.Offset, params.Limit))
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	response.JSON(ctx, w, http.StatusOK, mapAPITags(tags))
}

func (t TaskHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	var body api.CreateTagJSONBody
	if err := request.Bind(r, &body); err != nil {
		response.Error(ctx, w, err)

		return
	}

	tag, err := t.serv.AddTag(r.Context(), body.Name, userID)
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	response.JSON(ctx, w, http.StatusCreated, api.Tag{Id: tag.ID, Name: tag.Name})
}

func (t TaskHandler) GetTag(w http.ResponseWriter, r *http.Request, tagID int) {
	ctx := r.Context()
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	tag, err := t.serv.GetTag(r.Context(), tagID, userID)
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	response.JSON(ctx, w, http.StatusOK, api.Tag{Id: tag.ID, Name: tag.Name})
}

func (t TaskHandler) UpdateTag(w http.ResponseWriter, r *http.Request, tagID int) {
	ctx := r.Context()
	userID, err := authmiddleware.UserIDFromCtx(r.Context())
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	var body api.Tag
	err = request.Bind(r, &body)
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	err = t.serv.UpdateTag(r.Context(), tagID, body.Name, userID)
	if err != nil {
		response.Error(ctx, w, err)

		return
	}

	response.Status(w, http.StatusOK)
}
