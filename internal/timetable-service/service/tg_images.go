package service

import (
	"context"
	"fmt"

	"github.com/Dyleme/Notifier/internal/timetable-service/domains"
)

//go:generate mockgen -destination=mocks/tg_images_mocks.go -package=mocks . TgImagesRepository
type TgImagesRepository interface {
	Add(ctx context.Context, filename, tgFileID string) error
	Get(ctx context.Context, filename string) (domains.TgImage, error)
}

func (s *Service) GetTgImage(ctx context.Context, filename string) (domains.TgImage, error) {
	op := "Service.GetTgImage: %w"

	tgImage, err := s.repo.TgImages().Get(ctx, filename)
	if err != nil {
		err = fmt.Errorf(op, err)
		logError(ctx, err)

		return domains.TgImage{}, err
	}

	return tgImage, nil
}

func (s *Service) AddTgImage(ctx context.Context, filename, tgFileID string) error {
	op := "Service.AddTgImage: %w"

	err := s.repo.TgImages().Add(ctx, filename, tgFileID)
	if err != nil {
		err = fmt.Errorf(op, err)
		logError(ctx, err)

		return err
	}

	return nil
}
