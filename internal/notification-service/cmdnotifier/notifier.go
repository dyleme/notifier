package cmdnotifier

import (
	"context"
	"log"

	"golang.org/x/exp/slog"

	"github.com/Dyleme/Notifier/internal/timetable-service/models"
)

type Notifier struct {
	logger *slog.Logger
}

func (n Notifier) Notify(_ context.Context, ns models.SendingNotification) error {
	log.Printf("%+v\n", ns)
	return nil
}

func New(logger *slog.Logger) *Notifier {
	return &Notifier{logger: logger}
}
