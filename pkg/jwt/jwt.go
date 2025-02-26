package jwt

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

var ErrInvalidToken = errors.New("invalid token")

// Claims описывает полезную нагрузку JWT-токена.
type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	SecretKey string
}

// NewJWTManager создаёт новый экземпляр менеджера JWT.
func NewJWTManager(secretKey string) *JWTManager {
	return &JWTManager{
		SecretKey: secretKey,
	}
}

// VerifyToken проверяет JWT-токен, полученный от сервиса регистрации.
func (j *JWTManager) VerifyToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.SecretKey), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid || claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
