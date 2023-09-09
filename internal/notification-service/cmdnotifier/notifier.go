package cmdnotifier

import (
	"context"
	"log"
	"log/slog"

	"github.com/Dyleme/Notifier/internal/timetable-service/domains"
)

type Notifier struct {
	logger *slog.Logger
}

func (n Notifier) Notify(_ context.Context, ns domains.SendingNotification) error {
	log.Printf("%+v\n", ns)
	return nil
}

func New(logger *slog.Logger) *Notifier {
	return &Notifier{logger: logger}
}
