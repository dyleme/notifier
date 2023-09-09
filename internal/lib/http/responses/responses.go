package responses

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Dyleme/Notifier/internal/lib/log"
	"github.com/Dyleme/Notifier/internal/lib/serverrors"
)

type errorResponse struct {
	Err string `json:"error_message"`
}

var errServer = errors.New("server error")

func KnownError(w http.ResponseWriter, err error) {
	unwrappedErr := errors.Unwrap(err)
	if unwrappedErr != nil {
		err = unwrappedErr
	}

	if _, ok := err.(serverrors.InternalError); ok { //nolint:errorlint //error is already unwrapped
		Error(w, http.StatusInternalServerError, errServer)
	} else {
		switch err.(type) { //nolint:errorlint //error is already unwrapped
		case serverrors.NotFoundError:
			Error(w, http.StatusNotFound, err)
		case serverrors.NoDeletionsError:
			Error(w, http.StatusUnprocessableEntity, err)
		case serverrors.UniqueError:
			Error(w, http.StatusConflict, err)
		case serverrors.InvalidAuthError:
			Error(w, http.StatusUnauthorized, err)
		case serverrors.BusinessLogicError:
			Error(w, http.StatusUnprocessableEntity, err)
		default:
			Error(w, http.StatusInternalServerError, err)
		}
	}
}

func Error(w http.ResponseWriter, statusCode int, err error) {
	(w).Header().Set("Content-Type", "application/json; charset=utf-8")
	(w).Header().Set("X-Content-Type-Options", "nosniff")

	js, err := json.Marshal(errorResponse{err.Error()})

	if err != nil {
		log.Ctx(context.TODO()).Error(err.Error())
		statusCode = http.StatusInternalServerError
	}

	w.WriteHeader(statusCode)

	fmt.Fprint(w, string(js))
}

func JSON(w http.ResponseWriter, statusCode int, obj any) {
	js, err := json.Marshal(obj)
	if err != nil {
		Error(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	fmt.Fprint(w, string(js))
}

func Status(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)
}
