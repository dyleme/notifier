package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/dyleme/Notifier/internal/domain"
	"github.com/dyleme/Notifier/internal/domain/apperr"
)

//go:generate mockgen -destination=mocks/tg_images_mocks.go -package=mocks . TgImagesRepository
type TgImagesRepository interface {
	Add(ctx context.Context, filename, tgFileID string) error
	Get(ctx context.Context, filename string) (domain.TgImage, error)
}

func (s *Service) GetTgImage(ctx context.Context, filename string) (domain.TgImage, error) {
	tgImage, err := s.repos.tgImages.Get(ctx, filename)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return domain.TgImage{}, fmt.Errorf("get tg image: %w", apperr.NotFoundError{Object: "tg image"})
		}

		return domain.TgImage{}, err
	}

	return tgImage, nil
}

func (s *Service) AddTgImage(ctx context.Context, filename, tgFileID string) error {
	op := "Service.AddTgImage: %w"

	err := s.repos.tgImages.Add(ctx, filename, tgFileID)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}
