package service

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/Dyleme/Notifier/internal/authorization/models"
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
	input.Password = s.hashGen.GeneratePasswordHash(input.Password)

	user, err := s.repo.Create(ctx, input)
	if err != nil {
		return "", err
	}

	accessToken, err := s.jwtGen.CreateToken(user.ID)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

var ErrWrongPassword = errors.New("wrong password")

// AuthUser returns the jwt token of the user, if the provided user exists  in repo and password is correct.
// In any other situation function returns ("", err).
// Method get password and if calling repo.Get then validates it with the hashGen.IsValidPassword,
// and create token with the help jwtGen.CreateToken.
func (s *AuthService) AuthUser(ctx context.Context, input ValidateUserInput) (string, error) {
	user, err := s.repo.Get(ctx, input.AuthName, nil)
	if err != nil {
		return "", fmt.Errorf("get user: %w", err)
	}

	if !s.hashGen.IsValidPassword(input.Password, user.PasswordHash) {
		return "", ErrWrongPassword
	}

	return s.jwtGen.CreateToken(user.ID)
}
