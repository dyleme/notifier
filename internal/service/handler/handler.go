package handler

import (
	"github.com/Dyleme/Notifier/internal/service/service"
)

type EventHandler struct {
	serv *service.Service
}

func New(serv *service.Service) EventHandler {
	return EventHandler{serv: serv}
}
