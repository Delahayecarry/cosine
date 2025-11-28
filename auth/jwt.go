package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the JWT token claims
type JWTClaims struct {
	UserID            int64  `json:"user_id"`
	Username          string `json:"username"`
	Name              string `json:"name"`
	LinuxDoID         int    `json:"linuxdo_id"`
	LinuxDoTrustLevel int    `json:"trust_level"`
	jwt.RegisteredClaims
}

// GenerateToken generates a new JWT token for a LinuxDo user
func GenerateToken(userID int64, username, name string, linuxDoID, trustLevel int, secret string) (string, error) {
	claims := JWTClaims{
		UserID:            userID,
		Username:          username,
		Name:              name,
		LinuxDoID:         linuxDoID,
		LinuxDoTrustLevel: trustLevel,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * 7 * time.Hour)), // 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken parses and validates a JWT token
func ParseToken(tokenString, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
