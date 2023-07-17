package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/Dyleme/Notifier/internal/authorization/authmiddleware"
	"github.com/Dyleme/Notifier/internal/authorization/handler/authapi"
	authHandlerrs "github.com/Dyleme/Notifier/internal/authorization/handler/handlers"
	"github.com/Dyleme/Notifier/internal/authorization/jwt"
	authRepository "github.com/Dyleme/Notifier/internal/authorization/repository"
	authenticationService "github.com/Dyleme/Notifier/internal/authorization/service"
	"github.com/Dyleme/Notifier/internal/config"
	"github.com/Dyleme/Notifier/internal/lib/sqldatabase"
	timetableHandler "github.com/Dyleme/Notifier/internal/timetable-service/handler/handlers"
	"github.com/Dyleme/Notifier/internal/timetable-service/handler/timetableapi"
	timetableRepository "github.com/Dyleme/Notifier/internal/timetable-service/repository"
	timetableService "github.com/Dyleme/Notifier/internal/timetable-service/service"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()
	db, err := sqldatabase.NewPGX(ctx, cfg.Database.ConnectionString())
	if err != nil {
		log.Fatal(err)
	}

	timetableRepo := timetableRepository.New(db)
	serv := timetableService.New(timetableRepo)
	timeTableHandler := timetableHandler.New(serv)

	jwtGen := jwt.NewJwtGen(cfg.JWT)

	jwtMiddleware := authmiddleware.NewJWT(jwtGen)
	apiTokenMiddleware := authmiddleware.NewAPIToken(cfg.APIKey)
	authRepo := authRepository.New(db)
	authService := authenticationService.NewAuth(authRepo, &authenticationService.HashGen{}, jwtGen)
	authHandler := authHandlerrs.New(authService)

	router := chi.NewRouter()
	router.Use(
		middleware.RequestLogger(&middleware.DefaultLogFormatter{
			Logger:  &DefLogger{},
			NoColor: true,
		}),
		middleware.DefaultLogger,
		cors.AllowAll().Handler,
		middleware.Recoverer,
	)
	bearerTokenRouter := router.With(jwtMiddleware.Handle)
	apiKeyRouter := router.With(apiTokenMiddleware.Handle)

	fmt.Println(authHandler, apiTokenMiddleware)
	authapi.HandlerFromMux(authHandler, apiKeyRouter)
	timetableapi.HandlerFromMux(timeTableHandler, bearerTokenRouter)

	addr := fmt.Sprintf("localhost:%d", cfg.AppPort)
	err = http.ListenAndServe(addr, router) //nolint:gosec // TODO:add custom client
	if err != nil {
		log.Fatal(err)
	}
}

type DefLogger struct{}

func (dl *DefLogger) Print(vs ...any) {
	for _, v := range vs {
		fmt.Printf("%#v\n", v)
	}
}
