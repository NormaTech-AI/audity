package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// AuthMiddleware validates JWT tokens and injects user context
func AuthMiddleware(jwtSecret string, logger *zap.SugaredLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var tokenString string

			// Try to get token from Authorization header first
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" {
				// Check if it's a Bearer token
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenString = parts[1]
				}
			}

			// If no token in header, try to get from cookie
			if tokenString == "" {
				cookie, err := c.Cookie("auth_token")
				// logger.Info("Cookie", "cookie", cookie)
				if err == nil && cookie.Value != "" {
					tokenString = cookie.Value
				}
			}
			// logger.Info("Token string", "token_string", tokenString)
			// If still no token, return unauthorized
			if tokenString == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Missing authentication token",
				})
			}

			// Parse and validate token
			token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
				// Verify signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil {
				logger.Warnw("Failed to parse token", "error", err)
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid or expired token",
				})
			}

			claims, ok := token.Claims.(*JWTClaims)
			if !ok || !token.Valid {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid token claims",
				})
			}

			// Store claims in context for handlers to use
			c.Set("user", claims)
			c.Set("user_id", claims.UserID)
			c.Set("user_email", claims.Email)

			logger.Debugw("User authenticated", "user_id", claims.UserID, "email", claims.Email)

			return next(c)
		}
	}
}

// GetUserFromContext retrieves user claims from context
func GetUserFromContext(c echo.Context) (*JWTClaims, error) {
	user, ok := c.Get("user").(*JWTClaims)
	if !ok {
		return nil, fmt.Errorf("user not found in context")
	}
	return user, nil
}

// ClientPoolCache interface for getting client database pools
type ClientPoolCache interface {
	GetClientPool(ctx context.Context, clientID uuid.UUID) (*pgxpool.Pool, error)
}

// TenantStore interface for querying tenant database
type TenantStore interface {
	GetPool() *pgxpool.Pool
}

// AuthMiddlewareWithClientDB validates JWT tokens and injects both user context and client_db pool
// This middleware fetches the user's client_id from tenant_db and adds the client_db pool to context
func AuthMiddlewareWithClientDB(jwtSecret string, tenantStore TenantStore, clientPoolCache ClientPoolCache, logger *zap.SugaredLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var tokenString string

			// Try to get token from Authorization header first
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" {
				// Check if it's a Bearer token
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenString = parts[1]
				}
			}

			// If no token in header, try to get from cookie
			if tokenString == "" {
				cookie, err := c.Cookie("auth_token")
				if err == nil && cookie.Value != "" {
					tokenString = cookie.Value
				}
			}

			// If still no token, return unauthorized
			if tokenString == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Missing authentication token",
				})
			}

			// Parse and validate token
			token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
				// Verify signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil {
				logger.Warnw("Failed to parse token", "error", err)
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid or expired token",
				})
			}

			claims, ok := token.Claims.(*JWTClaims)
			if !ok || !token.Valid {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid token claims",
				})
			}

			// Store claims in context for handlers to use
			c.Set("user", claims)
			c.Set("user_id", claims.UserID)
			c.Set("user_email", claims.Email)

			// Fetch user's client_id from tenant_db
			ctx := c.Request().Context()
			var clientID *uuid.UUID
			query := `SELECT client_id FROM users WHERE id = $1`
			
			var clientIDBytes []byte
			err = tenantStore.GetPool().QueryRow(ctx, query, claims.UserID).Scan(&clientIDBytes)
			if err != nil {
				// User might not have a client_id (e.g., internal users)
				logger.Debugw("User has no client_id", "user_id", claims.UserID)
			} else if clientIDBytes != nil {
				parsedClientID, err := uuid.FromBytes(clientIDBytes)
				if err == nil {
					clientID = &parsedClientID
					c.Set("client_id", parsedClientID)
					
					// Get client database pool and add to context
					clientPool, err := clientPoolCache.GetClientPool(ctx, parsedClientID)
					if err != nil {
						logger.Errorw("Failed to get client database pool", 
							"error", err, 
							"client_id", parsedClientID,
							"user_id", claims.UserID)
						// Don't fail the request, just log the error
						// Some endpoints might not need client_db access
					} else {
						c.Set("client_db", clientPool)
						logger.Debugw("Client database pool added to context",
							"client_id", parsedClientID,
							"user_id", claims.UserID)
					}
				}
			}

			logger.Debugw("User authenticated", 
				"user_id", claims.UserID, 
				"email", claims.Email,
				"client_id", clientID)

			return next(c)
		}
	}
}

// GetClientDBFromContext retrieves client database pool from context
func GetClientDBFromContext(c echo.Context) (*pgxpool.Pool, error) {
	pool, ok := c.Get("client_db").(*pgxpool.Pool)
	if !ok {
		return nil, fmt.Errorf("client_db not found in context")
	}
	return pool, nil
}

// GetClientIDFromContext retrieves client ID from context
func GetClientIDFromContext(c echo.Context) (uuid.UUID, error) {
	clientID, ok := c.Get("client_id").(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("client_id not found in context")
	}
	return clientID, nil
}
