package server

import (
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"

	authApi "github.com/Dyleme/Notifier/internal/authorization/handler/api"
	tasksApi "github.com/Dyleme/Notifier/internal/service/handler/api"
)

func Route(
	timetableHandler tasksApi.ServerInterface,
	authHandler authApi.ServerInterface,
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

	authApi.HandlerFromMux(authHandler, apiKeyRouter)
	tasksApi.HandlerFromMux(timetableHandler, bearerTokenRouter)
	router.Get("/debug", pprof.Profile)

	return router
}
