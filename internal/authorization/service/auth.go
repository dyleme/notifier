package service

import (
	"context"
	"errors"
	"fmt"

	trManager "github.com/avito-tech/go-transaction-manager/trm/manager"
	"golang.org/x/crypto/bcrypt"

	"github.com/Dyleme/Notifier/internal/domain"
)

// HashGenerator interface providing you the ability to generate password hash
// and compare it with pure text passoword.
type HashGenerator interface {
	GeneratePasswordHash(password string) string
	IsValidPassword(password string, hash []byte) bool
}

type CodeSender interface {
	SendBindingMessage(ctx context.Context, tgID int, code string) error
}

// HashGen struct is realization of the HashGenerator interface with the bcrypt package.
type HashGen struct{}

// GeneratePasswordHash generates hash from the password.
func (h *HashGen) GeneratePasswordHash(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)

	return string(hash)
}

// IsValidPassword compare the password and password hash,
// returns true if they correspond, false in the other situations.
func (h *HashGen) IsValidPassword(password string, hash []byte) bool {
	err := bcrypt.CompareHashAndPassword(hash, []byte(password))

	return err == nil
}

type ValidateUserInput struct {
	AuthName string
	Password string
}

type BindingAttempt struct {
	ID           int
	TGID         int
	Code         string
	PasswordHash string
	Done         bool
}

// UserRepo is an interface which provides methods to implement with repository.
type UserRepo interface {
	Create(ctx context.Context, input CreateUserInput) (user domain.User, err error)
	Find(ctx context.Context, tgNickname string, tgID int) (domain.User, error)
	Update(ctx context.Context, user domain.User) error
}

// AuthService struct provides the ability to create user and validate it.
type AuthService struct {
	repo    UserRepo
	hashGen HashGenerator
	tg      CodeSender
	tr      *trManager.Manager
}

// NewAuth is the constructor to the AuthService.
func NewAuth(repo UserRepo, hashGen HashGenerator, tr *trManager.Manager) *AuthService {
	return &AuthService{
		repo:    repo,
		hashGen: hashGen,
		tr:      tr,
		tg:      nil,
	}
}

func (s *AuthService) SetCodeSender(tg CodeSender) {
	s.tg = tg
}

type StartUserBindingInput struct {
	TGNickname string
	Password   string
}

func (s *AuthService) GetTGUserInfo(ctx context.Context, tgID int) (domain.User, error) {
	op := "AuthService.GetTGUserInfo: %w"
	user, err := s.repo.Find(ctx, "", tgID)
	if err != nil {
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
			TGNickname: input.TGNickname,
			TGID:       input.TGID,
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
		user, err := s.repo.Find(ctx, "", id)
		if err != nil {
			return fmt.Errorf("get user: %w", err)
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
