package v1

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"CabBookingService/internal/controllers/helper"
	"CabBookingService/internal/repositories"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const (
	AccountKey contextKey = "account"
)

// AuthMiddleware verifies the JWT token and adds the user to the context
func AuthMiddleware(jwtSecret string, accountRepo repositories.AccountRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Get the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				helper.RespondWithError(w, http.StatusUnauthorized, "Authorization header required")
				return
			}

			// 2. Remove "Bearer " prefix
			tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

			// 3. Parse and validate the token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				helper.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			// 4. Get the User ID from claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				helper.RespondWithError(w, http.StatusUnauthorized, "Invalid token claims")
				return
			}

			// In our AuthService, "sub" is the Account ID (string)
			accountIDStr, ok := claims["sub"].(string)
			if !ok {
				helper.RespondWithError(w, http.StatusUnauthorized, "Invalid account ID in token")
				return
			}

			// 5. Fetch the account from the DB
			// We need to parse the UUID string first if your GetByID expects uuid.UUID
			uid, err := uuid.Parse(accountIDStr)
			if err != nil {
				helper.RespondWithError(w, http.StatusUnauthorized, "Invalid user ID format")
				return
			}

			account, err := accountRepo.GetByID(uid)
			if err != nil {
				helper.RespondWithError(w, http.StatusUnauthorized, "Account not found")
				return
			}

			// 6. Add to context
			ctx := context.WithValue(r.Context(), AccountKey, account)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
