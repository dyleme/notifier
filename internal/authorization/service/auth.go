package service

import (
	"context"
	"errors"
	"fmt"

	trManager "github.com/avito-tech/go-transaction-manager/trm/manager"
	"golang.org/x/crypto/bcrypt"

	"github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/pkg/serverrors"
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
	AddBindingAttempt(ctx context.Context, input BindingAttempt) error
	GetLatestBindingAttempt(ctx context.Context, tgID int) (BindingAttempt, error)
	UpdateBindingAttemptStatus(ctx context.Context, baID int, done bool) error
}

// JwtGenerator is an interface that provides method to create jwt tokens.
type JwtGenerator interface {
	// CreateToken is method which creates jwt token.
	CreateToken(id int) (string, error)
}

type CodeGenerator interface {
	GenereateCode() string
}

// AuthService struct provides the ability to create user and validate it.
type AuthService struct {
	repo          UserRepo
	codeGenerator CodeGenerator
	hashGen       HashGenerator
	jwtGen        JwtGenerator
	tg            CodeSender
	tr            *trManager.Manager
}

// NewAuth is the constructor to the AuthService.
func NewAuth(repo UserRepo, hashGen HashGenerator, jwtGen JwtGenerator, tr *trManager.Manager, codeGen CodeGenerator) *AuthService {
	return &AuthService{
		repo:          repo,
		codeGenerator: codeGen,
		hashGen:       hashGen,
		jwtGen:        jwtGen,
		tr:            tr,
		tg:            nil,
	}
}

func (s *AuthService) SetCodeSender(tg CodeSender) {
	s.tg = tg
}

type StartUserBindingInput struct {
	TGNickname string
	Password   string
}

// CreateUser function returns the id of the created user or error if any occures.
// Function get password hash of the user and creates user and calls CreateUser method of repository.
func (s *AuthService) StartUserBinding(ctx context.Context, input StartUserBindingInput) error {
	code := s.codeGenerator.GenereateCode()
	passwordHash := s.hashGen.GeneratePasswordHash(input.Password)
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		user, err := s.repo.Find(ctx, input.TGNickname, 0)
		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}
		err = s.repo.AddBindingAttempt(ctx, BindingAttempt{
			ID:           0,
			TGID:         user.TGID,
			Code:         code,
			PasswordHash: passwordHash,
			Done:         false,
		})
		if err != nil {
			return fmt.Errorf("add binding attempt: %w", err)
		}

		err = s.tg.SendBindingMessage(ctx, user.TGID, code)
		if err != nil {
			return fmt.Errorf("send binding message: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("tr: %w", err)
	}

	return nil
}

type BindUserInput struct {
	Code       string
	TGNickname string
}

func (s *AuthService) BindUser(ctx context.Context, input BindUserInput) error {
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		user, err := s.repo.Find(ctx, input.TGNickname, 0)
		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}

		ba, err := s.repo.GetLatestBindingAttempt(ctx, user.TGID)
		if err != nil {
			return fmt.Errorf("get binding attempt: %w", err)
		}

		if input.Code != ba.Code {
			return ErrWrongCode
		}

		user.PasswordHash = []byte(ba.PasswordHash)

		err = s.repo.Update(ctx, user)
		if err != nil {
			return fmt.Errorf("update user: %w", err)
		}

		err = s.repo.UpdateBindingAttemptStatus(ctx, ba.ID, true)
		if err != nil {
			return fmt.Errorf("update binding attempt status: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("tr: %w", err)
	}

	return nil
}

var (
	ErrWrongPassword = errors.New("wrong password")
	ErrWrongCode     = errors.New("wrong code")
)

// AuthUser returns the jwt token of the user, if the provided user exists  in repo and password is correct.
// In any other situation function returns ("", err).
// Method get password and if calling repo.Get then validates it with the hashGen.IsValidPassword,
// and create token with the help jwtGen.CreateToken.
func (s *AuthService) AuthUser(ctx context.Context, input ValidateUserInput) (string, error) {
	op := "AuthService.AuthUser: %w"
	user, err := s.repo.Find(ctx, input.AuthName, 0)
	if err != nil {
		return "", fmt.Errorf(op, err)
	}

	if !s.hashGen.IsValidPassword(input.Password, user.PasswordHash) {
		return "", ErrWrongPassword
	}

	token, err := s.jwtGen.CreateToken(user.ID)
	if err != nil {
		return "", fmt.Errorf(op, err)
	}

	return token, nil
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
	if !tzOffset.IsValid() {
		return fmt.Errorf("invalid offset: %w", serverrors.NewBusinessLogicError("invalid offset"))
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
