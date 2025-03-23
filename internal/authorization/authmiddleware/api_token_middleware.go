package authmiddleware

import (
	"errors"
	"net/http"

	"github.com/Dyleme/Notifier/internal/service/handler/response"
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
		ctx := r.Context()
		token, err := getAPIKey(r)
		if err != nil {
			response.Error(ctx, w, err)

			return
		}

		if token != am.apiToken {
			response.Error(ctx, w, ErrInvalidAuthKey)

			return
		}

		next.ServeHTTP(w, r)
	})
}

func getAPIKey(r *http.Request) (string, error) {
	apiKeys, exist := r.Header[apiKeyHeader]
	if !exist {
		return "", errors.New("no api key in request")
	}

	if len(apiKeys) != 1 {
		return "", errors.New("more than one auth header")
	}

	return apiKeys[0], nil
}
