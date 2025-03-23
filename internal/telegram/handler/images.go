package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	serverrors "github.com/Dyleme/Notifier/internal/domain/apperr"
)

func (th *TelegramHandler) SendImage(ctx context.Context, filename string, image []byte, sendPhotoParams *bot.SendPhotoParams) error {
	op := "TelegramHandler.SendImage: %w"

	photo, knownImage, err := th.getInputFile(ctx, filename, image)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	sendPhotoParams.Photo = photo

	res, err := th.bot.SendPhoto(ctx, sendPhotoParams)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	if !knownImage {
		err = th.serv.AddTgImage(ctx, filename, res.Photo[0].FileID)
		if err != nil {
			return fmt.Errorf(op, err)
		}
	}

	return nil
}

func (th *TelegramHandler) getInputFile(ctx context.Context, filename string, data []byte) (inputFile models.InputFile, known bool, err error) { //nolint:nonamedreturns // better readability
	op := "TelegramHandler.getInputFile: %w"
	tgImage, err := th.serv.GetTgImage(ctx, filename)
	if err != nil {
		var notFoundErr serverrors.NotFoundError
		if errors.As(err, &notFoundErr) {
			return &models.InputFileUpload{
				Filename: filename,
				Data:     bytes.NewReader(data),
			}, false, nil
		}

		return nil, false, fmt.Errorf(op, err)
	}

	return &models.InputFileString{
		Data: tgImage.TgFileID,
	}, true, nil
}
