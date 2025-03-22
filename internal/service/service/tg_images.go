package service

import (
	"context"
	"fmt"

	"github.com/Dyleme/Notifier/internal/domain"
)

//go:generate mockgen -destination=mocks/tg_images_mocks.go -package=mocks . TgImagesRepository
type TgImagesRepository interface {
	Add(ctx context.Context, filename, tgFileID string) error
	Get(ctx context.Context, filename string) (domain.TgImage, error)
}

func (s *Service) GetTgImage(ctx context.Context, filename string) (domain.TgImage, error) {
	op := "Service.GetTgImage: %w"

	tgImage, err := s.repos.tgImages.Get(ctx, filename)
	if err != nil {
		err = fmt.Errorf(op, err)
		logError(ctx, err)

		return domain.TgImage{}, err
	}

	return tgImage, nil
}

func (s *Service) AddTgImage(ctx context.Context, filename, tgFileID string) error {
	op := "Service.AddTgImage: %w"

	err := s.repos.tgImages.Add(ctx, filename, tgFileID)
	if err != nil {
		err = fmt.Errorf(op, err)
		logError(ctx, err)

		return err
	}

	return nil
}
