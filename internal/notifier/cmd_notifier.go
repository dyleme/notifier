package notifier

import (
	"context"

	"github.com/dyleme/Notifier/internal/domain"
	"github.com/dyleme/Notifier/pkg/log"
)

type CmdNotifier struct{}

func (cn CmdNotifier) Notify(ctx context.Context, notif domain.Event) error {
	log.Ctx(ctx).Info("notify", "notification", notif)

	return nil
}
