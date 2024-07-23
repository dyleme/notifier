package service

import (
	"context"

	"github.com/Dyleme/Notifier/pkg/log"
)

type CmdCodeSender struct{}

func (s CmdCodeSender) SendBindingMessage(ctx context.Context, tgID int, code string) error {
	log.Ctx(ctx).Info("SendBindingMessage", "tgID", tgID, "code", code)

	return nil
}
