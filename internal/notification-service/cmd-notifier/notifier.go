package notifier

import (
	"context"
	"log"

	"github.com/Dyleme/Notifier/internal/notification-service/notifier/notification"
)

type Notifier struct{}

func (n Notifier) Notify(_ context.Context, ns []notification.Notification) error {
	for _, n := range ns {
		log.Printf("%+v\n", n)
	}
	return nil
}

func New() *Notifier {
	return &Notifier{}
}
