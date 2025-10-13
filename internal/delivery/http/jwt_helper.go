package http

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

func NewJWTManager(secret string, durationHours int) *JWTManager {
	return &JWTManager{
		secretKey:     secret,
		tokenDuration: time.Duration(durationHours) * time.Hour,
	}
}

func (j *JWTManager) GenerateToken(userID uint64, role, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"email":   email,
		"exp":     time.Now().Add(j.tokenDuration).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}
