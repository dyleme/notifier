package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/internal/domain/apperr"
	"github.com/Dyleme/Notifier/pkg/database/txmanager"
)

// UserRepo is an interface which provides methods to implement with repository.
type UserRepo interface {
	Get(ctx context.Context, id int) (domain.User, error)
	Create(ctx context.Context, input CreateUserInput) (user domain.User, err error)
	FindByTgID(ctx context.Context, tgID int) (domain.User, error)
	Update(ctx context.Context, user domain.User) error
}

// AuthService struct provides the ability to create user and validate it.
type AuthService struct {
	repo UserRepo
	tr   *txmanager.TxManager
}

// NewAuth is the constructor to the AuthService.
func NewAuth(repo UserRepo, tr *txmanager.TxManager) *AuthService {
	return &AuthService{
		repo: repo,
		tr:   tr,
	}
}

func (s *AuthService) GetTGUserInfo(ctx context.Context, tgID int) (domain.User, error) {
	op := "AuthService.GetTGUserInfo: %w"
	user, err := s.repo.FindByTgID(ctx, tgID)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return domain.User{}, apperr.NotFoundError{Object: "user"}
		}
		return domain.User{}, fmt.Errorf(op, err)
	}

	return user, nil
}

type CreateUserInput struct {
	TGNickname string
	TGID       int
}

func (s *AuthService) CreateUser(ctx context.Context, input CreateUserInput) (domain.User, error) {
	var user domain.User
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var err error
		user, err = s.repo.Create(ctx, CreateUserInput{
			TGNickname: input.TGNickname, TGID: input.TGID,
		})
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

func (s *AuthService) UpdateUserTime(ctx context.Context, id int, tzOffset domain.TimeZoneOffset, isDst bool) error {
	if err := tzOffset.Valid(); err != nil {
		return fmt.Errorf("invalid offset: %w", err)
	}

	err := s.tr.Do(ctx, func(ctx context.Context) error {
		user, err := s.repo.Get(ctx, id)
		if err != nil {
			return fmt.Errorf("get user id[%v]: %w", id, err)
		}
		user.IsTimeZoneDST = isDst
		user.TimeZoneOffset = int(tzOffset)

		err = s.repo.Update(ctx, user)
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
