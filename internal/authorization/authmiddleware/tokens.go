package authmiddleware

import (
	"errors"
	"net/http"
	"strings"
)

const (
	authorizationHeader string = "Authorization"
)

type authorizationToken string

func getAuthorizationTokens(r *http.Request, token authorizationToken) (string, error) {
	authHeader, exist := r.Header[authorizationHeader]
	if !exist {
		return "", errors.New("empty auth header")
	}

	if len(authHeader) != 1 {
		return "", errors.New("more than one auth header")
	}

	auth := authHeader[0]

	if auth[:len(token)] != string(token) {
		return "", errors.New("invalid authentication method")
	}

	authJWT := auth[len(token):]
	authJWT = strings.TrimPrefix(authJWT, " ")

	return authJWT, nil
}
