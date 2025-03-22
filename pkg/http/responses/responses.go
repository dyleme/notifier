package responses

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Dyleme/Notifier/internal/domain/apperr"
	"github.com/Dyleme/Notifier/pkg/log"
)

type errorResponse struct {
	ErrorMessage string `json:"error_message"`
}

type BadBodyResponse struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

var errServer = errors.New("server error")

func KnownError(w http.ResponseWriter, err error) {
	unwrappedErr := errors.Unwrap(err)
	if unwrappedErr != nil {
		unwrappedErr = err
	}

	if _, ok := unwrappedErr.(apperr.InternalError); ok { //nolint:errorlint //error is already unwrapped
		Error(w, http.StatusInternalServerError, errServer)

		return
	}

	if mappingError, ok := unwrappedErr.(apperr.MappingError); ok { //nolint:errorlint //error is already unwrapped
		MapError(w, BadBodyResponse{
			Field: mappingError.Field,
			Error: mappingError.Error(),
		})

		return
	}

	switch unwrappedErr.(type) { //nolint:errorlint //error is already unwrapped
	case apperr.NotFoundError:
		Error(w, http.StatusNotFound, unwrappedErr)
	case apperr.NoDeletionsError:
		Error(w, http.StatusUnprocessableEntity, unwrappedErr)
	case apperr.UniqueError:
		Error(w, http.StatusConflict, unwrappedErr)
	case apperr.InvalidAuthError:
		Error(w, http.StatusUnauthorized, unwrappedErr)
	case apperr.BusinessLogicError:
		Error(w, http.StatusUnprocessableEntity, unwrappedErr)
	default:
		Error(w, http.StatusInternalServerError, unwrappedErr)
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

func MapError(w http.ResponseWriter, badBodyResp BadBodyResponse) {
	(w).Header().Set("Content-Type", "application/json; charset=utf-8")
	(w).Header().Set("X-Content-Type-Options", "nosniff")

	statusCode := http.StatusBadRequest
	js, err := json.Marshal(badBodyResp)
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
