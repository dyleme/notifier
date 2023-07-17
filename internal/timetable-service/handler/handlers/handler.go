package handlers

import (
	"github.com/Dyleme/Notifier/internal/timetable-service/service"
)

type TimetableHandler struct {
	serv *service.Service
}

func New(serv *service.Service) TimetableHandler {
	return TimetableHandler{serv: serv}
}
