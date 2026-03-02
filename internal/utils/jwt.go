package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTCustomClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

func GenerateTokenPair(userID uuid.UUID, role, secret string) (string, string, error) {
	// 1. Access Token (Short-lived)
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, JWTCustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})
	at, err := accessToken.SignedString([]byte(secret))
	if err != nil {
		return "", "", err
	}

	// 2. Refresh Token (Long-lived - Opaque or JWT)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		Subject:   userID.String(),
	})
	rt, err := refreshToken.SignedString([]byte(secret))

	return at, rt, err
}
