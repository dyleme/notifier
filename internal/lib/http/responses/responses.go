package responses

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type errorResponse struct {
	Err string `json:"error_message"`
}

func Error(w http.ResponseWriter, statusCode int, err error) {
	(w).Header().Set("Content-Type", "application/json; charset=utf-8")
	(w).Header().Set("X-Content-Type-Options", "nosniff")

	js, err := json.Marshal(errorResponse{err.Error()})

	if err != nil {
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
