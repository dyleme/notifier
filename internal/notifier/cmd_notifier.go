package notifier

import (
	"context"

	"github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/pkg/log"
)

type CmdNotifier struct{}

func (cn CmdNotifier) Notify(ctx context.Context, notif domain.Notification) error {
	log.Ctx(ctx).Info("notify", "notification", notif)

	return nil
}
