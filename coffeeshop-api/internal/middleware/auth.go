package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const (
	// UserIDKey is the context key for the authenticated user's ID.
	UserIDKey contextKey = "user_id"
	// UserRoleKey is the context key for the authenticated user's role.
	UserRoleKey contextKey = "user_role"
	// TenantIDKey is the context key for the authenticated user's tenant.
	TenantIDKey contextKey = "tenant_id"
	// DeviceIDKey is the context key for the POS device making the request.
	DeviceIDKey contextKey = "device_id"
)

// Auth validates the JWT token from the Authorization header and injects
// the user ID, role, and tenant ID into the request context.
func Auth(jwtSecret string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing authorization header","code":401}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				http.Error(w, `{"error":"invalid authorization header format","code":401}`, http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, `{"error":"invalid or expired token","code":401}`, http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, `{"error":"invalid token claims","code":401}`, http.StatusUnauthorized)
				return
			}

			userID, _ := claims["sub"].(string)
			role, _ := claims["role"].(string)
			tenantID, _ := claims["tenant_id"].(string)

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, UserRoleKey, role)
			ctx = context.WithValue(ctx, TenantIDKey, tenantID)

			// Read optional X-Device-ID header from POS terminals
			if deviceID := r.Header.Get("X-Device-ID"); deviceID != "" {
				ctx = context.WithValue(ctx, DeviceIDKey, deviceID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

// GetUserID extracts the user ID from the request context.
func GetUserID(ctx context.Context) string {
	id, _ := ctx.Value(UserIDKey).(string)
	return id
}

// GetUserRole extracts the user role from the request context.
func GetUserRole(ctx context.Context) string {
	role, _ := ctx.Value(UserRoleKey).(string)
	return role
}

// GetTenantID extracts the tenant ID from the request context as a UUID.
func GetTenantID(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(TenantIDKey).(string)
	parsed, _ := uuid.Parse(id)
	return parsed
}

// GetDeviceID extracts the device ID from the request context as a UUID pointer.
// Returns nil if no X-Device-ID header was provided.
func GetDeviceID(ctx context.Context) *uuid.UUID {
	id, _ := ctx.Value(DeviceIDKey).(string)
	if id == "" {
		return nil
	}
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil
	}
	return &parsed
}
