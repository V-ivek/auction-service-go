package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"auction-microservice/pkg/config"
)

// JWTMiddleware provides JWT authentication middleware
type JWTMiddleware struct {
	config *config.Config
	logger *zap.Logger
}

// NewJWTMiddleware creates a new JWT middleware instance
func NewJWTMiddleware(config *config.Config, logger *zap.Logger) *JWTMiddleware {
	return &JWTMiddleware{
		config: config,
		logger: logger,
	}
}

// Authenticate is a middleware that validates JWT tokens
func (m *JWTMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		token := m.extractToken(r)
		if token == "" {
			m.writeUnauthorized(w, "missing or invalid authorization token")
			return
		}

		// Validate token and extract user ID
		userID, err := m.ValidateToken(token)
		if err != nil {
			m.logger.Debug("token validation failed", zap.Error(err))
			m.writeUnauthorized(w, "invalid token")
			return
		}

		// Add user ID to request context
		ctx := context.WithValue(r.Context(), "user_id", userID)
		
		// Add additional claims to context if needed
		ctx = context.WithValue(ctx, "authenticated", true)
		ctx = context.WithValue(ctx, "auth_time", time.Now())

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ValidateToken validates a JWT token and returns the user ID
func (m *JWTMiddleware) ValidateToken(tokenString string) (uuid.UUID, error) {
	// For now, we'll implement a simple token validation
	// In production, you would use a proper JWT library like github.com/golang-jwt/jwt
	
	// Simple token format: "user:{uuid}"
	// This is a mock implementation for development/testing
	if !strings.HasPrefix(tokenString, "user:") {
		return uuid.Nil, fmt.Errorf("invalid token format")
	}

	userIDStr := strings.TrimPrefix(tokenString, "user:")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID in token: %w", err)
	}

	m.logger.Debug("token validated successfully", zap.String("user_id", userID.String()))
	return userID, nil
}

// GenerateToken generates a JWT token for a user
func (m *JWTMiddleware) GenerateToken(userID uuid.UUID) (string, error) {
	// Simple token format for development/testing
	// In production, use a proper JWT library
	token := fmt.Sprintf("user:%s", userID.String())
	
	m.logger.Debug("token generated", zap.String("user_id", userID.String()))
	return token, nil
}

// extractToken extracts the token from the Authorization header
func (m *JWTMiddleware) extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Expected format: "Bearer {token}"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// writeUnauthorized writes an unauthorized response
func (m *JWTMiddleware) writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	
	response := fmt.Sprintf(`{"error": "unauthorized", "message": "%s"}`, message)
	w.Write([]byte(response))
}

// RequireAuth is a helper function that can be used to protect individual handlers
func (m *JWTMiddleware) RequireAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.Authenticate(http.HandlerFunc(handler)).ServeHTTP(w, r)
	}
}

// GetUserID extracts the user ID from the request context
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userID := ctx.Value("user_id")
	if userID == nil {
		return uuid.Nil, false
	}

	switch v := userID.(type) {
	case uuid.UUID:
		return v, true
	case string:
		if parsed, err := uuid.Parse(v); err == nil {
			return parsed, true
		}
	}

	return uuid.Nil, false
}

// IsAuthenticated checks if the request is authenticated
func IsAuthenticated(ctx context.Context) bool {
	authenticated := ctx.Value("authenticated")
	if authenticated == nil {
		return false
	}

	if auth, ok := authenticated.(bool); ok {
		return auth
	}

	return false
}

// GetAuthTime returns the authentication time from context
func GetAuthTime(ctx context.Context) (time.Time, bool) {
	authTime := ctx.Value("auth_time")
	if authTime == nil {
		return time.Time{}, false
	}

	if t, ok := authTime.(time.Time); ok {
		return t, true
	}

	return time.Time{}, false
}

// Optional middleware that allows both authenticated and unauthenticated requests
func (m *JWTMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		token := m.extractToken(r)
		
		ctx := r.Context()
		
		if token != "" {
			// If token is present, try to validate it
			if userID, err := m.ValidateToken(token); err == nil {
				// Token is valid, add user info to context
				ctx = context.WithValue(ctx, "user_id", userID)
				ctx = context.WithValue(ctx, "authenticated", true)
				ctx = context.WithValue(ctx, "auth_time", time.Now())
			} else {
				// Token is invalid, but we don't reject the request
				m.logger.Debug("optional auth: token validation failed", zap.Error(err))
				ctx = context.WithValue(ctx, "authenticated", false)
			}
		} else {
			// No token present, mark as unauthenticated
			ctx = context.WithValue(ctx, "authenticated", false)
		}

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}