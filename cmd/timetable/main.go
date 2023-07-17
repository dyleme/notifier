package main

import (
	"context"
	_ "embed"
	"log"

	"github.com/Dyleme/Notifier/internal/authorization/authmiddleware"
	authHandlerrs "github.com/Dyleme/Notifier/internal/authorization/handler/handlers"
	"github.com/Dyleme/Notifier/internal/authorization/jwt"
	authRepository "github.com/Dyleme/Notifier/internal/authorization/repository"
	authenticationService "github.com/Dyleme/Notifier/internal/authorization/service"
	"github.com/Dyleme/Notifier/internal/config"
	"github.com/Dyleme/Notifier/internal/lib/sqldatabase"
	"github.com/Dyleme/Notifier/internal/server"
	timetableHandler "github.com/Dyleme/Notifier/internal/timetable-service/handler/handlers"
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
	timetableServ := timetableService.New(timetableRepo)
	timeTableHandler := timetableHandler.New(timetableServ)

	apiTokenMiddleware := authmiddleware.NewAPIToken(cfg.APIKey)
	jwtGen := jwt.NewJwtGen(cfg.JWT)
	jwtMiddleware := authmiddleware.NewJWT(jwtGen)
	authRepo := authRepository.New(db)
	authService := authenticationService.NewAuth(authRepo, &authenticationService.HashGen{}, jwtGen)
	authHandler := authHandlerrs.New(authService)

	router := server.Route(
		timeTableHandler,
		authHandler,
		jwtMiddleware.Handle,
		apiTokenMiddleware.Handle,
	)

	serv := server.New(router, cfg.Server)
	if err := serv.Run(ctx); err != nil {
		log.Println(err)
	}
}
