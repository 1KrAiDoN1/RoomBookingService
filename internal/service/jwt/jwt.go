package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}
type JWTManager struct {
	jwtSecretKey []byte
	tokenTTL     time.Duration
}

func NewJWTManager(secretKey string, tokenTTL time.Duration) *JWTManager {
	return &JWTManager{
		jwtSecretKey: []byte(secretKey),
		tokenTTL:     tokenTTL,
	}
}

func (m *JWTManager) GenerateToken(userID string, role string) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.tokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.jwtSecretKey)
}

func (m *JWTManager) ParseToken(tokenStr string) (string, string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(m.jwtSecretKey), nil
	})

	if err != nil {
		return "", "", fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return "", "", fmt.Errorf("invalid token claims")
	}

	return claims.UserID, claims.Role, nil
}
