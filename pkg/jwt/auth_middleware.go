package jwt

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	// ContextUserIDKey используется для хранения ID пользователя в контексте.
	ContextUserIDKey contextKey = "userID"
)

// AuthMiddleware возвращает middleware, которое проверяет JWT-токен.
func AuthMiddleware(jwtManager *JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header missing", http.StatusUnauthorized)
				return
			}

			// Ожидаем формат: "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
				return
			}

			tokenStr := parts[1]
			claims, err := jwtManager.VerifyToken(tokenStr)
			if err != nil {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Добавляем ID пользователя в контекст
			ctx := context.WithValue(r.Context(), ContextUserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
