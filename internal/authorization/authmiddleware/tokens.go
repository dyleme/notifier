package authmiddleware

import (
	"fmt"
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
		return "", fmt.Errorf("empty auth header")
	}

	if len(authHeader) != 1 {
		return "", fmt.Errorf("more than one auth header")
	}

	auth := authHeader[0]

	if auth[:len(token)] != string(token) {
		return "", fmt.Errorf("invalid authentication method")
	}

	authJWT := auth[len(token):]
	authJWT = strings.TrimPrefix(authJWT, " ")

	return authJWT, nil
}
