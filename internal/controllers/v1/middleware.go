package v1

import (
	"context"
	"fmt"
	"net/http"

	"CabBookingService/internal/controllers/helper"
	"CabBookingService/internal/models"
	"CabBookingService/internal/repositories"
	"CabBookingService/internal/services"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang-jwt/jwt/v5/request"
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
			// 1. Get the bearer token from Authorization header using the library extractor
			// This handles "Bearer " prefix stripping automatically and robustly
			bearerToken, err := request.AuthorizationHeaderExtractor.ExtractToken(r)
			if err != nil {
				helper.RespondWithError(w, http.StatusUnauthorized, "Authorization header required: "+err.Error())
				return
			}

			// 2. Parse and validate the token into UserClaims
			claims := &services.UserClaims{}

			// jwt.ParseWithClaims automatically checks exp, nbf, iat if present in RegisteredClaims
			token, err := jwt.ParseWithClaims(bearerToken, claims, func(token *jwt.Token) (interface{}, error) {
				// Validate the signing method is what we expect
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil {
				// We can handle specific JWT errors here if needed (e.g., expired vs invalid)
				helper.RespondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}

			if !token.Valid {
				helper.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			// 4. Get the User ID from Subject claim
			// RegisteredClaims.Subject maps to the "sub" JSON field
			accountIDStr := claims.Subject
			if accountIDStr == "" {
				helper.RespondWithError(w, http.StatusUnauthorized, "Invalid token: missing subject")
				return
			}

			// 5. Fetch the account from the DB
			// We need to parse the UUID string first if your GetByID expects uuid.UUID
			uid, err := uuid.Parse(accountIDStr)
			if err != nil {
				helper.RespondWithError(w, http.StatusUnauthorized, "Invalid user ID format in token")
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
			account, err := GetAccountFromContext(r.Context())
			if err != nil {
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

// GetAccountFromContext is a helper to safely retrieve the typed account
func GetAccountFromContext(ctx context.Context) (*models.Account, error) {
	val := ctx.Value(AccountKey)
	if val == nil {
		return nil, fmt.Errorf("account not found in context")
	}
	account, ok := val.(*models.Account)
	if !ok {
		return nil, fmt.Errorf("invalid account type in context")
	}
	return account, nil
}
