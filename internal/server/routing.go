package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Dyleme/Notifier/internal/authorization/handler/authapi"
	"github.com/Dyleme/Notifier/internal/timetable-service/handler/timetableapi"
)

func Route(
	timetableHandler timetableapi.ServerInterface,
	authHandler authapi.ServerInterface,
	jwtMiddleware func(handler http.Handler) http.Handler,
	apiKeyMiddleware func(handler http.Handler) http.Handler,
	middlewares []func(handler http.Handler) http.Handler,
) *chi.Mux {
	router := chi.NewRouter()
	for _, m := range middlewares {
		router.Use(m)
	}
	bearerTokenRouter := router.With(jwtMiddleware)
	apiKeyRouter := router.With(apiKeyMiddleware)

	authapi.HandlerFromMux(authHandler, apiKeyRouter)
	timetableapi.HandlerFromMux(timetableHandler, bearerTokenRouter)

	return router
}
