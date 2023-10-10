package responses

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Dyleme/Notifier/pkg/log"
	"github.com/Dyleme/Notifier/pkg/serverrors"
)

type errorResponse struct {
	ErrorMessage string `json:"error_message"`
}

var errServer = errors.New("server error")

func KnownError(w http.ResponseWriter, err error) {
	unwrappedErr := errors.Unwrap(err)
	if unwrappedErr != nil {
		unwrappedErr = err
	}

	if _, ok := unwrappedErr.(serverrors.InternalError); ok { //nolint:errorlint //error is already unwrapped
		Error(w, http.StatusInternalServerError, errServer)
	} else {
		switch unwrappedErr.(type) { //nolint:errorlint //error is already unwrapped
		case serverrors.NotFoundError:
			Error(w, http.StatusNotFound, unwrappedErr)
		case serverrors.NoDeletionsError:
			Error(w, http.StatusUnprocessableEntity, unwrappedErr)
		case serverrors.UniqueError:
			Error(w, http.StatusConflict, unwrappedErr)
		case serverrors.InvalidAuthError:
			Error(w, http.StatusUnauthorized, unwrappedErr)
		case serverrors.BusinessLogicError:
			Error(w, http.StatusUnprocessableEntity, unwrappedErr)
		default:
			Error(w, http.StatusInternalServerError, unwrappedErr)
		}
	}
}

func Error(w http.ResponseWriter, statusCode int, err error) {
	(w).Header().Set("Content-Type", "application/json; charset=utf-8")
	(w).Header().Set("X-Content-Type-Options", "nosniff")

	js, err := json.Marshal(errorResponse{err.Error()})
	if err != nil {
		log.Default().Error(err.Error())
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
