package notifier

import (
	"context"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/pkg/log"
)

type CmdNotifier struct{}

func (cn CmdNotifier) Notify(ctx context.Context, notif domains.Notification) error {
	log.Ctx(ctx).Info("notify", "notification", notif)

	return nil
}
