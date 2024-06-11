package handler

import (
	"github.com/Dyleme/Notifier/internal/service/service"
)

type TaskHandler struct {
	serv *service.Service
}

func New(serv *service.Service) TaskHandler {
	return TaskHandler{serv: serv}
}
