package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dyleme/Notifier/internal/domain"
	"github.com/dyleme/Notifier/internal/domain/apperr"
)

// UserRepo is an interface which provides methods to implement with repository.
type UserRepo interface {
	Get(ctx context.Context, id int) (domain.User, error)
	Create(ctx context.Context, user domain.User) (domain.User, error)
	GetByTgID(ctx context.Context, tgID int) (domain.User, error)
	Update(ctx context.Context, user domain.User) error
}

func (s *Service) GetTGUser(ctx context.Context, tgID int) (domain.User, error) {
	op := "AuthService.GetTGUserInfo: %w"
	user, err := s.repos.users.GetByTgID(ctx, tgID)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return domain.User{}, apperr.NotFoundError{Object: "user"}
		}

		return domain.User{}, fmt.Errorf(op, err)
	}

	return user, nil
}

func (s *Service) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var err error
		user, err = s.repos.users.Create(ctx, user)
		if err != nil {
			return fmt.Errorf("create user: %w", err)
		}

		return nil
	})
	if err != nil {
		return domain.User{}, fmt.Errorf("tr: %w", err)
	}

	return user, nil
}

var ErrInvalidOffset = errors.New("invalid offset")

func (s *Service) UpdateUserTime(ctx context.Context, tgID int, tzOffset domain.TimeZoneOffset, isDst bool) error {
	if err := tzOffset.Valid(); err != nil {
		return fmt.Errorf("invalid offset: %w", err)
	}

	err := s.tr.Do(ctx, func(ctx context.Context) error {
		user, err := s.repos.users.GetByTgID(ctx, tgID)
		if err != nil {
			return fmt.Errorf("get user id[%v]: %w", tgID, err)
		}
		user.IsTimeZoneDST = isDst
		user.TimeZoneOffset = int(tzOffset)

		err = s.repos.users.Update(ctx, user)
		if err != nil {
			return fmt.Errorf("update user: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("tr: %w", err)
	}

	return nil
}

func (s *Service) SetDefaultNotificatinPeriod(ctx context.Context, period time.Duration, userID int) error {
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		user, err := s.repos.users.Get(ctx, userID)
		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}

		user.DefaultNotificationPeriod = period
		err = s.repos.users.Update(ctx, user)
		if err != nil {
			return fmt.Errorf("update user: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("tr: %w", err)
	}

	return nil
}
