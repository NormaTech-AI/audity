package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
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
