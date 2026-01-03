package v1

import (
	"CabBookingService/internal/controllers/helper"
	"CabBookingService/internal/models"
	"CabBookingService/internal/repositories"
	"context"
	"fmt"
	"net/http"
	"strings"

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

			account, err := accountRepo.GetByID(r.Context(), uid)
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

// RequireRoleMiddleware checks if the user has the required role
func RequireRoleMiddleware(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Get account from context
			account, ok := r.Context().Value(AccountKey).(*models.Account)
			if !ok || account == nil {
				helper.RespondWithError(w, http.StatusUnauthorized, "Account not found in context")
				return
			}

			// 2. Check if the account has the required role
			if !hasRole(account.Roles, requiredRole) {
				helper.RespondWithError(w, http.StatusForbidden, "Insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func hasRole(roles []models.Role, role string) bool {
	for _, r := range roles {
		if r.Name == role {
			return true
		}
	}
	return false
}
