package authmiddleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Dyleme/Notifier/pkg/http/responses"
)

type APITokenMiddleware struct {
	apiToken string
}

func NewAPIToken(apiToken string) APITokenMiddleware {
	return APITokenMiddleware{apiToken: apiToken}
}

const apiKeyHeader string = "Api_key"

var ErrInvalidAuthKey = errors.New("invalid auth key")

func (am *APITokenMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := getAPIKey(r)
		if err != nil {
			responses.Error(w, http.StatusUnauthorized, err)

			return
		}

		if token != am.apiToken {
			responses.Error(w, http.StatusUnauthorized, ErrInvalidAuthKey)

			return
		}

		next.ServeHTTP(w, r)
	})
}

func getAPIKey(r *http.Request) (string, error) {
	apiKeys, exist := r.Header[apiKeyHeader]
	if !exist {
		return "", fmt.Errorf("no api key in request")
	}

	if len(apiKeys) != 1 {
		return "", fmt.Errorf("more than one auth header")
	}

	return apiKeys[0], nil
}
