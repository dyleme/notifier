package service

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

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
	NickName string
	Email    string
	Password string
}

type ValidateUserInput struct {
	AuthName string
	Password string
}

// AuthRepo is an interface which provides methods to implement with repository.
type AuthRepo interface {
	// CreateUser creates user in the repository.
	CreateUser(ctx context.Context, input CreateUserInput) (id int, err error)

	// GetPasswordHashAndID returns user password hash and id.
	GetPasswordHashAndID(ctx context.Context, authName string) (hash []byte, userID int, err error)
}

// JwtGenerator is an interface that provides method to create jwt tokens.
type JwtGenerator interface {
	// CreateToken is method which creates jwt token.
	CreateToken(id int) (string, error)
}

// AuthService struct provides the ability to create user and validate it.
type AuthService struct {
	repo    AuthRepo
	hashGen HashGenerator
	jwtGen  JwtGenerator
}

// NewAuth is the constructor to the AuthService.
func NewAuth(repo AuthRepo, hashGen HashGenerator, jwtGen JwtGenerator) *AuthService {
	return &AuthService{repo: repo, hashGen: hashGen, jwtGen: jwtGen}
}

// CreateUser function returns the id of the created user or error if any occures.
// Function get password hash of the user and creates user and calls CreateUser method of repository.
func (s *AuthService) CreateUser(ctx context.Context, user CreateUserInput) (string, error) {
	user.Password = s.hashGen.GeneratePasswordHash(user.Password)

	id, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return "", err
	}

	accessToken, err := s.jwtGen.CreateToken(id)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

// AuthUser returns the jwt token of the user, if the provided user exists  in repo and password is correct.
// In any other situation function returns ("", err).
// Method get password and if calling repo.GetPasswordHashAndID then validates it with the hashGen.IsValidPassword,
// and create token with the help jwtGen.CreateToken.
func (s *AuthService) AuthUser(ctx context.Context, input ValidateUserInput) (string, error) {
	hash, id, err := s.repo.GetPasswordHashAndID(ctx, input.AuthName)
	if err != nil {
		return "", fmt.Errorf("get user: %w", err)
	}

	if !s.hashGen.IsValidPassword(input.Password, hash) {
		return "", serverrors.NewInvalidAuth("wrong password")
	}

	return s.jwtGen.CreateToken(id)
}
