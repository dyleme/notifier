package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

type Gen struct {
	signedKey string
	ttl       time.Duration
}

type Config struct {
	SignedKey string
	TTL       time.Duration
}

func NewJwtGen(config *Config) *Gen {
	return &Gen{signedKey: config.SignedKey, ttl: config.TTL}
}

var (
	ErrInvalidToken           = errors.New("invalid token")
	ErrTokenClaimsInvalidType = errors.New("token claims are not of the type MapClaims")
	ErrInvalidUserID          = errors.New("invalid userID in token")
)

type UnexpectedSingingMethodError struct {
	method interface{}
}

func (err UnexpectedSingingMethodError) Error() string {
	return fmt.Sprintf("unexpected singing method: %v", err.method)
}

type tokenClaims struct {
	jwt.Claims
	UserID int `json:"user_id"`
}

// CreateToken function generate token with provided TTL and user id.
func (g *Gen) CreateToken(userID int) (string, error) {
	op := "Gen.CreateToken: %w"
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		Claims: jwt.StandardClaims{ //nolint:exhaustruct // no need to fill
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(g.ttl).Unix(),
		},
		UserID: userID,
	})

	token, err := jwtToken.SignedString([]byte(g.signedKey))
	if err != nil {
		return "", fmt.Errorf(op, err)
	}

	return token, nil
}

// ParseToken function returns user id from JWT token, if this token is liquid.
func (g *Gen) ParseToken(tokenString string) (userID int, err error) { //nolint:nonamedreturns //better reading
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return 0, UnexpectedSingingMethodError{t.Header["alg"]}
		}

		return []byte(g.signedKey), nil
	})
	if err != nil {
		return 0, fmt.Errorf("parse token: %w", err)
	}

	if !token.Valid {
		return 0, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, ErrTokenClaimsInvalidType
	}

	userID, err = getUserID(claims)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func getUserID(claims jwt.MapClaims) (int, error) {
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("can't cust userID claims to flat64")
	}
	userID := int(userIDFloat)
	if !ok || userID == 0 {
		return 0, ErrInvalidUserID
	}

	return userID, nil
}
