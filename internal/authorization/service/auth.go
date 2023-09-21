package service

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/Dyleme/Notifier/internal/authorization/models"
	"github.com/Dyleme/Notifier/internal/lib/serverrors"
)

// HashGenerator interface providing you the ability to generate password hash
// and compare it with pure text passoword.
type HashGenerator interface {
	GeneratePasswordHash(password string) string
	IsValidPassword(password string, hash []byte) bool
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

type CreateUserInput struct {
	Email    string
	Password string
	TGID     *int
}

type ValidateUserInput struct {
	AuthName string
	Password string
}

// UserRepo is an interface which provides methods to implement with repository.
type UserRepo interface {
	Atomic(context.Context, func(ctx context.Context, repo UserRepo) error) error
	// CreateUser creates user in the repository.
	Create(ctx context.Context, input CreateUserInput) (user models.User, err error)

	// GetPasswordHashAndID returns user password hash and id.
	Get(ctx context.Context, authName string, tgID *int) (models.User, error)

	UpdateTime(ctx context.Context, id int, tzOffset models.TimeZoneOffset, isDST bool) error
}

type NotifcationService interface {
	CreateDefaultNotificationParams()
}

// JwtGenerator is an interface that provides method to create jwt tokens.
type JwtGenerator interface {
	// CreateToken is method which creates jwt token.
	CreateToken(id int) (string, error)
}

// AuthService struct provides the ability to create user and validate it.
type AuthService struct {
	repo    UserRepo
	hashGen HashGenerator
	jwtGen  JwtGenerator
}

// NewAuth is the constructor to the AuthService.
func NewAuth(repo UserRepo, hashGen HashGenerator, jwtGen JwtGenerator) *AuthService {
	return &AuthService{repo: repo, hashGen: hashGen, jwtGen: jwtGen}
}

// CreateUser function returns the id of the created user or error if any occures.
// Function get password hash of the user and creates user and calls CreateUser method of repository.
func (s *AuthService) CreateUser(ctx context.Context, input CreateUserInput) (string, error) {
	op := "AuthService.CreateUser: %w"
	input.Password = s.hashGen.GeneratePasswordHash(input.Password)

	user, err := s.repo.Create(ctx, input)
	if err != nil {
		return "", fmt.Errorf(op, err)
	}

	accessToken, err := s.jwtGen.CreateToken(user.ID)
	if err != nil {
		return "", fmt.Errorf(op, err)
	}

	return accessToken, nil
}

var ErrWrongPassword = errors.New("wrong password")

// AuthUser returns the jwt token of the user, if the provided user exists  in repo and password is correct.
// In any other situation function returns ("", err).
// Method get password and if calling repo.Get then validates it with the hashGen.IsValidPassword,
// and create token with the help jwtGen.CreateToken.
func (s *AuthService) AuthUser(ctx context.Context, input ValidateUserInput) (string, error) {
	op := "AuthService.AuthUser: %w"
	user, err := s.repo.Get(ctx, input.AuthName, nil)
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

func (s *AuthService) GetTGUserInfo(ctx context.Context, tgID int) (models.User, error) {
	op := "AuthService.GetTGUserInfo: %w"
	var tgUser models.User
	err := s.repo.Atomic(ctx, func(ctx context.Context, userRepo UserRepo) error {
		user, err := userRepo.Get(ctx, "", &tgID)
		if err == nil { // err equal nil
			tgUser = user
		}

		var notFoundErr serverrors.NotFoundError
		if !errors.As(err, &notFoundErr) {
			return err //nolint:wrapcheck // wrapping later
		}

		user, err = userRepo.Create(ctx, CreateUserInput{
			Email:    "",
			Password: "",
			TGID:     &tgID,
		})
		if err != nil {
			return err //nolint:wrapcheck // wrapping later
		}
		tgUser = user

		return nil
	})
	if err != nil {
		return models.User{}, fmt.Errorf(op, err)
	}

	return tgUser, nil
}

var ErrInvalidOffset = errors.New("invalid offset")

func (s *AuthService) UpdateUserTime(ctx context.Context, id int, tzOffset models.TimeZoneOffset, isDst bool) error {
	op := "AuthService.UpdateUserTime: %w"

	if !tzOffset.IsValid() {
		return fmt.Errorf(op, ErrInvalidOffset)
	}

	err := s.repo.Atomic(ctx, func(ctx context.Context, repo UserRepo) error {
		err := repo.UpdateTime(ctx, id, tzOffset, isDst)
		if err != nil {
			return err //nolint:wrapcheck //wrapping later
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}
