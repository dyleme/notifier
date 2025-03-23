package response

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Dyleme/Notifier/internal/domain/apperr"
	"github.com/Dyleme/Notifier/pkg/log"
)

type errorResp struct {
	HTTPCode int            `json:"-"`
	Code     string         `json:"code"`
	Message  string         `json:"message"`
	Details  map[string]any `json:"details,omitempty"`
}

func Error(ctx context.Context, w http.ResponseWriter, err error) {
	(w).Header().Set("Content-Type", "application/json; charset=utf-8")
	(w).Header().Set("X-Content-Type-Options", "nosniff")

	resp, ok := check(err,
		checkValidation,
		checkBusiness,
	)
	if !ok {
		log.Ctx(ctx).Error("unhandled error", log.Err(err))
		resp = errorResp{
			HTTPCode: http.StatusInternalServerError,
			Code:     "INTERNAL_ERROR",
			Message:  "Internal server error. Try againg later",
			Details:  nil,
		}
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Ctx(ctx).Error("marshal error", log.Err(err))
	}

	w.WriteHeader(resp.HTTPCode)
}

func check(err error, fs ...func(err error) (errorResp, bool)) (errorResp, bool) {
	for _, f := range fs {
		resp, ok := f(err)
		if ok {
			return resp, ok
		}
	}

	return errorResp{}, false
}

func checkValidation(err error) (errorResp, bool) {
	var resp errorResp

	var (
		jsonErr    *json.UnmarshalTypeError
		syntaxErr  *json.SyntaxError
		parsingErr apperr.ParsingError
	)
	switch {
	case errors.As(err, &jsonErr):
		resp = errorResp{
			HTTPCode: http.StatusBadRequest,
			Code:     "BAD_REQUEST",
			Message:  "Cannot be parsed",
			Details: map[string]any{
				"field": jsonErr.Field,
				"type":  jsonErr.Type,
				"value": jsonErr.Value,
			},
		}
	case errors.As(err, &syntaxErr):
		resp = errorResp{
			HTTPCode: http.StatusBadRequest,
			Code:     "BAD_REQUEST",
			Message:  "Invalid json",
			Details: map[string]any{
				"offset": syntaxErr.Offset,
			},
		}
	case errors.As(err, &parsingErr):
		resp = errorResp{
			HTTPCode: http.StatusBadRequest,
			Code:     "BAD_REQUEST",
			Message:  "Cannot be parsed",
			Details: map[string]any{
				"field": parsingErr.Field,
				"cause": parsingErr.Cause,
			},
		}
	default:
		return errorResp{}, false
	}

	return resp, true
}

func checkBusiness(err error) (errorResp, bool) {
	var resp errorResp

	var (
		notFoundErr        apperr.NotFoundError
		uniqueErr          apperr.UniqueError
		notBelongToUserErr apperr.NotBelongToUserError
	)
	switch {
	case errors.Is(err, apperr.ErrEventPastType):
		resp = errorResp{
			HTTPCode: http.StatusBadRequest,
			Code:     "BAD_REQUEST",
			Message:  "Event type should not be in the past. Use time in the future",
			Details:  map[string]any{},
		}
	case errors.As(err, &notFoundErr):
		resp = errorResp{
			HTTPCode: http.StatusNotFound,
			Code:     "NOT_FOUND",
			Message:  "object not found",
			Details: map[string]any{
				"object": notFoundErr.Object,
			},
		}
	case errors.As(err, &uniqueErr):
		resp = errorResp{
			HTTPCode: http.StatusConflict,
			Code:     "CONFLICT",
			Message:  "object already exists",
			Details: map[string]any{
				"conflicted": uniqueErr.Name,
				"value":      uniqueErr.Value,
			},
		}
	case errors.As(err, &notBelongToUserErr):
		resp = errorResp{
			HTTPCode: http.StatusForbidden,
			Code:     "FORBIDDEN",
			Message:  "Does not belongs to you",
			Details: map[string]any{
				"object":    notBelongToUserErr.ObjType,
				"object_id": notBelongToUserErr.ObjID,
			},
		}
	default:
		return errorResp{}, false
	}

	return resp, true
}

func JSON(ctx context.Context, w http.ResponseWriter, statusCode int, obj any) {
	js, err := json.Marshal(obj)
	if err != nil {
		Error(ctx, w, err)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	fmt.Fprint(w, string(js))
}

func Status(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)
}
