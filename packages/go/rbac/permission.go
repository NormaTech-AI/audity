package rbac

import (
	"context"
	"fmt"
	"net/http"

	"github.com/NormaTech-AI/audity/packages/go/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Store interface for database operations
type Store interface {
	GetPool() *pgxpool.Pool
}

// PermissionMiddleware checks if user has required permission
func PermissionMiddleware(store Store, logger *zap.SugaredLogger, requiredPermission string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get user from context (set by AuthMiddleware)
			user, err := auth.GetUserFromContext(c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "User not authenticated",
				})
			}

			// Check if user has the required permission
			hasPermission, err := checkUserPermission(c.Request().Context(), store, user.UserID, requiredPermission)
			if err != nil {
				logger.Errorw("Failed to check permission", "error", err, "user_id", user.UserID, "permission", requiredPermission)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Failed to verify permissions",
				})
			}

			if !hasPermission {
				logger.Warnw("Permission denied", "user_id", user.UserID, "permission", requiredPermission)
				return c.JSON(http.StatusForbidden, map[string]string{
					"error":    "Insufficient permissions",
					"required": requiredPermission,
				})
			}

			logger.Debugw("Permission granted", "user_id", user.UserID, "permission", requiredPermission)

			return next(c)
		}
	}
}

// RequirePermissions creates middleware that requires multiple permissions (AND logic)
func RequirePermissions(store Store, logger *zap.SugaredLogger, permissions ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, err := auth.GetUserFromContext(c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "User not authenticated",
				})
			}

			// Check all required permissions
			for _, permission := range permissions {
				hasPermission, err := checkUserPermission(c.Request().Context(), store, user.UserID, permission)
				if err != nil {
					logger.Errorw("Failed to check permission", "error", err, "user_id", user.UserID, "permission", permission)
					return c.JSON(http.StatusInternalServerError, map[string]string{
						"error": "Failed to verify permissions",
					})
				}

				if !hasPermission {
					logger.Warnw("Permission denied", "user_id", user.UserID, "permission", permission)
					return c.JSON(http.StatusForbidden, map[string]string{
						"error":    "Insufficient permissions",
						"required": permission,
					})
				}
			}

			return next(c)
		}
	}
}

// RequireAnyPermission creates middleware that requires at least one permission (OR logic)
func RequireAnyPermission(store Store, logger *zap.SugaredLogger, permissions ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, err := auth.GetUserFromContext(c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "User not authenticated",
				})
			}

			// Check if user has at least one of the required permissions
			for _, permission := range permissions {
				hasPermission, err := checkUserPermission(c.Request().Context(), store, user.UserID, permission)
				if err != nil {
					logger.Errorw("Failed to check permission", "error", err, "user_id", user.UserID, "permission", permission)
					continue
				}

				if hasPermission {
					logger.Debugw("Permission granted", "user_id", user.UserID, "permission", permission)
					return next(c)
				}
			}

			logger.Warnw("No required permissions found", "user_id", user.UserID, "permissions", permissions)
			return c.JSON(http.StatusForbidden, map[string]string{
				"error":    "Insufficient permissions",
				"required": "One of: " + joinPermissions(permissions),
			})
		}
	}
}

// RequireRole creates middleware that requires a specific role
// Fetches role from database, not from JWT
func RequireRole(store Store, logger *zap.SugaredLogger, requiredRole string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, err := auth.GetUserFromContext(c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "User not authenticated",
				})
			}

			// Fetch user's role from database
			userRole, err := getUserRole(c.Request().Context(), store, user.UserID)
			if err != nil {
				logger.Errorw("Failed to fetch user role", "error", err, "user_id", user.UserID)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Failed to verify user role",
				})
			}

			if userRole != requiredRole {
				logger.Warnw("Role mismatch", "user_id", user.UserID, "user_role", userRole, "required_role", requiredRole)
				return c.JSON(http.StatusForbidden, map[string]string{
					"error":         "Insufficient permissions",
					"required_role": requiredRole,
				})
			}

			return next(c)
		}
	}
}

// RequireAnyRole creates middleware that requires one of multiple roles
// Fetches role from database, not from JWT
func RequireAnyRole(store Store, logger *zap.SugaredLogger, roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, err := auth.GetUserFromContext(c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "User not authenticated",
				})
			}

			// Fetch user's role from database
			userRole, err := getUserRole(c.Request().Context(), store, user.UserID)
			if err != nil {
				logger.Errorw("Failed to fetch user role", "error", err, "user_id", user.UserID)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Failed to verify user role",
				})
			}

			for _, role := range roles {
				if userRole == role {
					return next(c)
				}
			}

			logger.Warnw("No matching role", "user_id", user.UserID, "user_role", userRole, "required_roles", roles)
			return c.JSON(http.StatusForbidden, map[string]string{
				"error":         "Insufficient permissions",
				"required_role": "One of: " + joinPermissions(roles),
			})
		}
	}
}

// getUserRole fetches the user's role from the database
func getUserRole(ctx context.Context, store Store, userID uuid.UUID) (string, error) {
	query := `SELECT role FROM users WHERE id = $1`

	var role string
	err := store.GetPool().QueryRow(ctx, query, userID).Scan(&role)
	if err != nil {
		return "", err
	}

	return role, nil
}

// checkUserPermission checks if a user has a specific permission
func checkUserPermission(ctx context.Context, store Store, userID uuid.UUID, permissionName string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM permissions p
			JOIN role_permissions rp ON p.id = rp.permission_id
			JOIN user_roles ur ON rp.role_id = ur.role_id
			WHERE ur.user_id = $1 AND p.name = $2
		) AS has_permission
	`

	var hasPermission bool
	err := store.GetPool().QueryRow(ctx, query, userID, permissionName).Scan(&hasPermission)
	if err != nil {
		return false, err
	}

	return hasPermission, nil
}

// joinPermissions joins permission names with commas
func joinPermissions(permissions []string) string {
	result := ""
	for i, p := range permissions {
		if i > 0 {
			result += ", "
		}
		result += p
	}
	return result
}

// ClientPermissionMiddleware checks if user has required permission in their client database
// This uses the client_db pool from context (set by AuthMiddlewareWithClientDB)
func ClientPermissionMiddleware(logger *zap.SugaredLogger, requiredPermission string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get user from context (set by AuthMiddleware)
			user, err := auth.GetUserFromContext(c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "User not authenticated",
				})
			}

			// Get client_db pool from context
			clientDB, err := auth.GetClientDBFromContext(c)
			if err != nil {
				logger.Errorw("Client database not found in context", 
					"error", err, 
					"user_id", user.UserID)
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "Client database access required",
				})
			}

			// Check if user has the required permission in client_db
			hasPermission, err := checkClientUserPermission(c.Request().Context(), clientDB, user.UserID, requiredPermission)
			if err != nil {
				logger.Errorw("Failed to check client permission", 
					"error", err, 
					"user_id", user.UserID, 
					"permission", requiredPermission)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Failed to verify permissions",
				})
			}

			if !hasPermission {
				logger.Warnw("Client permission denied", 
					"user_id", user.UserID, 
					"permission", requiredPermission)
				return c.JSON(http.StatusForbidden, map[string]string{
					"error":    "Insufficient permissions",
					"required": requiredPermission,
				})
			}

			logger.Debugw("Client permission granted", 
				"user_id", user.UserID, 
				"permission", requiredPermission)

			return next(c)
		}
	}
}

// checkClientUserPermission checks if a user has a specific permission in their client database
func checkClientUserPermission(ctx context.Context, clientDB *pgxpool.Pool, tenantUserID uuid.UUID, permissionName string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM client_permissions p
			JOIN client_role_permissions rp ON p.id = rp.permission_id
			JOIN client_user_roles ur ON rp.role_id = ur.role_id
			JOIN client_users cu ON ur.user_id = cu.id
			WHERE cu.tenant_user_id = $1 AND p.name = $2 AND cu.is_active = true
		) AS has_permission
	`

	var hasPermission bool
	err := clientDB.QueryRow(ctx, query, tenantUserID, permissionName).Scan(&hasPermission)
	if err != nil {
		return false, fmt.Errorf("failed to check client permission: %w", err)
	}

	return hasPermission, nil
}

// GetClientUserRole fetches the user's role from their client database
func GetClientUserRole(ctx context.Context, clientDB *pgxpool.Pool, tenantUserID uuid.UUID) (string, error) {
	query := `SELECT role FROM client_users WHERE tenant_user_id = $1 AND is_active = true`

	var role string
	err := clientDB.QueryRow(ctx, query, tenantUserID).Scan(&role)
	if err != nil {
		return "", fmt.Errorf("failed to fetch client user role: %w", err)
	}

	return role, nil
}

// ClientRequireAnyPermission creates middleware that requires at least one permission in client_db (OR logic)
func ClientRequireAnyPermission(logger *zap.SugaredLogger, permissions ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, err := auth.GetUserFromContext(c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "User not authenticated",
				})
			}

			// Get client_db pool from context
			clientDB, err := auth.GetClientDBFromContext(c)
			if err != nil {
				logger.Errorw("Client database not found in context", 
					"error", err, 
					"user_id", user.UserID)
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "Client database access required",
				})
			}

			// Check if user has at least one of the required permissions
			for _, permission := range permissions {
				hasPermission, err := checkClientUserPermission(c.Request().Context(), clientDB, user.UserID, permission)
				if err != nil {
					logger.Errorw("Failed to check client permission", 
						"error", err, 
						"user_id", user.UserID, 
						"permission", permission)
					continue
				}

				if hasPermission {
					logger.Debugw("Client permission granted", 
						"user_id", user.UserID, 
						"permission", permission)
					return next(c)
				}
			}

			logger.Warnw("No required client permissions found", 
				"user_id", user.UserID, 
				"permissions", permissions)
			return c.JSON(http.StatusForbidden, map[string]string{
				"error":    "Insufficient permissions",
				"required": "One of: " + joinPermissions(permissions),
			})
		}
	}
}
