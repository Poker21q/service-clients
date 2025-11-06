package jwtauth

import (
	"context"
	"net/http"
	"strings"

	"service-boilerplate-go/internal/pkg/response"

	"github.com/golang-jwt/jwt/v5"
)

type key int

const userIDKey key = iota

// Middleware проверяет JWT токен
func Middleware(jwtSecret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				response.ErrorStatus(w, http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrTokenMalformed
				}
				jwtSecretBytes := []byte(jwtSecret)
				return jwtSecretBytes, nil
			})
			if err != nil || !token.Valid {
				response.ErrorStatus(w, http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				response.ErrorStatus(w, http.StatusUnauthorized)
				return
			}

			userIDStr, ok := claims["user_id"].(string)
			if !ok {
				response.ErrorStatus(w, http.StatusUnauthorized)
				return
			}

			// Кладём userID в контекст
			ctx := context.WithValue(r.Context(), userIDKey, userIDStr)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext достаёт userID из контекста
func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}
