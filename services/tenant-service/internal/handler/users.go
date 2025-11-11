package handler

import (
	"net/http"

	"github.com/NormaTech-AI/audity/services/tenant-service/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

// ============================================================================
// Request/Response Types
// ============================================================================

type UserResponse struct {
	ID           string  `json:"id"`
	Email        string  `json:"email"`
	Name         string  `json:"name"`
	OIDCProvider string  `json:"oidc_provider"`
	Designation         string  `json:"designation"`
	ClientID     *string `json:"client_id"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
	LastLogin    *string `json:"last_login"`
}

type ListUsersResponse struct {
	Data []UserResponse `json:"data"`
}

// ============================================================================
// Handlers
// ============================================================================

// ListUsers lists all users, optionally filtered by client_id
// @Summary List users
// @Tags users
// @Produce json
// @Param client_id query string false "Filter by client ID"
// @Success 200 {object} ListUsersResponse
// @Router /api/users [get]
func (h *Handler) ListUsers(c echo.Context) error {
	clientIDParam := c.QueryParam("client_id")
	// roleParam := c.QueryParam("role")
	var users []db.User
	var err error
	h.logger.Info("Listing users", "client_id", clientIDParam)
	if clientIDParam == "tenant"{
		users, err = h.store.Queries.ListTenantUsers(c.Request().Context() )
	} else if clientIDParam != "" {
		// Filter by client_id
		clientID, parseErr := uuid.Parse(clientIDParam)
		if parseErr != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid client_id format")
		}
		users, err = h.store.Queries.ListUsersByClient(c.Request().Context(), pgtype.UUID{
			Bytes: clientID,
			Valid: true,
		})
	} else {
		// List all users
		users, err = h.store.Queries.ListUsers(c.Request().Context())
	}

	if err != nil {
		h.logger.Errorw("Failed to list users", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list users")
	}

	response := ListUsersResponse{
		Data: make([]UserResponse, len(users)),
	}

	for i, user := range users {
		userResp := UserResponse{
			ID:           user.ID.String(),
			Email:        user.Email,
			Name:         user.Name,
			OIDCProvider: user.OidcProvider,
			Designation:         string(user.Designation),
			CreatedAt:    user.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:    user.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		}

		if user.ClientID.Valid {
			clientUUID := uuid.UUID(user.ClientID.Bytes)
			clientID := clientUUID.String()
			userResp.ClientID = &clientID
		}

		if user.LastLogin.Valid {
			lastLogin := user.LastLogin.Time.Format("2006-01-02T15:04:05Z07:00")
			userResp.LastLogin = &lastLogin
		}

		response.Data[i] = userResp
	}

	return c.JSON(http.StatusOK, response)
}

// GetUser gets a user by ID
// @Summary Get user by ID
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} UserResponse
// @Router /api/users/{id} [get]
func (h *Handler) GetUser(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	user, err := h.store.Queries.GetUser(c.Request().Context(), userID)
	if err != nil {
		h.logger.Errorw("Failed to get user", "error", err, "user_id", userID)
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	response := UserResponse{
		ID:           user.ID.String(),
		Email:        user.Email,
		Name:         user.Name,
		OIDCProvider: user.OidcProvider,
		Designation:         string(user.Designation),
		CreatedAt:    user.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    user.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}

	if user.ClientID.Valid {
		clientUUID := uuid.UUID(user.ClientID.Bytes)
		clientID := clientUUID.String()
		response.ClientID = &clientID
	}

	if user.LastLogin.Valid {
		lastLogin := user.LastLogin.Time.Format("2006-01-02T15:04:05Z07:00")
		response.LastLogin = &lastLogin
	}

	return c.JSON(http.StatusOK, response)
}
