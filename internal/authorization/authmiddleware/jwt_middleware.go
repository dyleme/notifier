package authmiddleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Dyleme/Notifier/internal/authorization/jwt"
	"github.com/Dyleme/Notifier/internal/lib/http/responses"
)

type Key string

const (
	keyUserID Key = "keyUserID"
)

const (
	bearerToken authorizationToken = "Bearer"
)

var (
	ErrNoUserIDInCtx = errors.New("no user id in context")
)

type JWTMiddleware struct {
	gen *jwt.Gen
}

func NewJWT(gen *jwt.Gen) JWTMiddleware {
	return JWTMiddleware{gen: gen}
}

type JWTGen interface {
	CreateToken(userID int) (string, error)
	ParseToken(tokenString string) (userID int, err error)
}

func (am *JWTMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		authJWT, err := getAuthorizationTokens(r, bearerToken)
		if err != nil {
			responses.Error(w, http.StatusUnauthorized, fmt.Errorf("jwt token: %w", err))
			return
		}

		userID, err := am.gen.ParseToken(authJWT)
		if err != nil {
			responses.Error(w, http.StatusUnauthorized, fmt.Errorf("middleware: %w", err))
			return
		}

		ctx = storeUserID(ctx, userID)

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func storeUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, keyUserID, userID)
}

func UserIDFromCtx(ctx context.Context) (int, error) {
	userID, ok := ctx.Value(keyUserID).(int)
	if !ok {
		return 0, ErrNoUserIDInCtx
	}

	return userID, nil
}
