package handlers

import (
	"github.com/Dyleme/Notifier/internal/timetable-service/service"
)

type EventHandler struct {
	serv *service.Service
}

func New(serv *service.Service) EventHandler {
	return EventHandler{serv: serv}
}
