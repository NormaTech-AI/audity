package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/NormaTech-AI/audity/packages/go/auth"
	"github.com/NormaTech-AI/audity/services/auth-service/internal/db"
	"github.com/NormaTech-AI/audity/services/auth-service/internal/oidc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

// LoginRequest represents a login request
type LoginRequest struct {
	Provider string `json:"provider" validate:"required,oneof=google microsoft"`
}

// LoginResponse contains the OAuth URL to redirect to
type LoginResponse struct {
	AuthURL  string `json:"auth_url"`
	Provider string `json:"provider"`
}

// CallbackResponse contains the JWT token after successful authentication
type CallbackResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserInfo  `json:"user"`
}

// UserInfo represents user information in the response
type UserInfo struct {
	ID       uuid.UUID  `json:"id"`
	Email    string     `json:"email"`
	Name     string     `json:"name"`
	Role     string     `json:"role"`
	ClientID *uuid.UUID `json:"client_id,omitempty"`
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	Token string `json:"token" validate:"required"`
}

// InitiateLogin godoc
// @Summary Initiate OAuth login
// @Description Get OAuth URL for Google or Microsoft login
// @Tags auth
// @Accept json
// @Produce json
// @Param provider path string true "OAuth provider (google or microsoft)"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]string
// @Router /auth/login/{provider} [get]
func (h *Handler) InitiateLogin(c echo.Context) error {
	provider := c.Param("provider")

	var oidcProvider *oidc.OIDCProvider
	switch provider {
	case "google":
		oidcProvider = h.googleProvider
	case "microsoft":
		oidcProvider = h.microsoftProvider
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid provider. Must be 'google' or 'microsoft'",
		})
	}

	// Generate state for CSRF protection
	state, err := oidc.GenerateState()
	if err != nil {
		h.logger.Errorw("Failed to generate state", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate state",
		})
	}

	// Store state (in production, use Redis with expiration)
	h.stateStore[state] = provider

	// Get OAuth URL
	authURL := oidcProvider.GetAuthURL(state)

	return c.JSON(http.StatusOK, LoginResponse{
		AuthURL:  authURL,
		Provider: provider,
	})
}

// HandleCallback godoc
// @Summary OAuth callback handler
// @Description Handle OAuth callback and issue JWT token
// @Tags auth
// @Produce json
// @Param code query string true "Authorization code"
// @Param state query string true "State parameter"
// @Success 200 {object} CallbackResponse
// @Failure 400 {object} map[string]string
// @Router /auth/callback [get]
func (h *Handler) HandleCallback(c echo.Context) error {
	code := c.QueryParam("code")
	state := c.QueryParam("state")

	if code == "" || state == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing code or state parameter",
		})
	}

	// Verify state
	provider, exists := h.stateStore[state]
	if !exists {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid state parameter",
		})
	}
	delete(h.stateStore, state) // Remove used state

	// Get OIDC provider
	var oidcProvider *oidc.OIDCProvider
	switch provider {
	case "google":
		oidcProvider = h.googleProvider
	case "microsoft":
		oidcProvider = h.microsoftProvider
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid provider",
		})
	}

	ctx := c.Request().Context()

	// Exchange code for token
	token, err := oidcProvider.ExchangeCode(ctx, code)
	if err != nil {
		h.logger.Errorw("Failed to exchange code", "error", err, "provider", provider)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to exchange authorization code",
		})
	}

	// Get user info from provider
	userInfo, err := oidcProvider.GetUserInfo(ctx, token)
	if err != nil {
		h.logger.Errorw("Failed to get user info", "error", err, "provider", provider)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get user information",
		})
	}
	h.logger.Info("USER info by OIDC", userInfo)
	// Find or create user in database
	user, err := h.findOrCreateUser(ctx, userInfo, provider)
	if err != nil {
		h.logger.Errorw("Failed to find or create user", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to process user",
		})
	}
	h.logger.Info("User found or created", "user_id", user.ID, "email", user.Email)
	// Generate JWT token (only user_id and email - other data fetched from DB)
	jwtToken, err := h.jwtManager.GenerateToken(user.ID, user.Email)
	if err != nil {
		h.logger.Errorw("Failed to generate JWT", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate token",
		})
	}

	// Update last login
	if err := h.updateLastLogin(ctx, user.ID); err != nil {
		h.logger.Warnw("Failed to update last login", "error", err, "user_id", user.ID)
	}

	// Set token as HTTP-only cookie
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    jwtToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(h.config.Auth.JWTExpirationHours * 3600),
	}
	c.SetCookie(cookie)

	h.logger.Infow("Cookie set in callback",
		"cookie_name", cookie.Name,
		"cookie_path", cookie.Path,
		"cookie_http_only", cookie.HttpOnly,
		"redirect_url", h.config.Auth.FrontendURL,
	)

	// Redirect to frontend (token is now in cookie, no need to pass in URL)
	redirectURL := fmt.Sprintf("%s/auth/callback?success=true", h.config.Auth.FrontendURL)
	return c.Redirect(http.StatusFound, redirectURL)
}

// SetTokenCookie godoc
// @Summary Set JWT token as HTTP-only cookie
// @Description Accept JWT token from URL and set it as an HTTP-only cookie
// @Tags auth
// @Accept json
// @Produce json
// @Param token query string true "JWT token"
// @Success 200 {object} UserInfo
// @Failure 400 {object} map[string]string
// @Router /auth/set-token [post]
func (h *Handler) SetTokenCookie(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		h.logger.Warn("SetTokenCookie: Missing token parameter")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing token parameter",
		})
	}

	// Log token length for debugging (don't log the actual token)
	h.logger.Debugw("SetTokenCookie: Received token", "token_length", len(token))

	// Validate the token
	claims, err := h.jwtManager.ValidateToken(token)
	if err != nil {
		h.logger.Warnw("SetTokenCookie: Token validation failed",
			"error", err,
			"token_length", len(token),
			"token_prefix", token[:min(10, len(token))],
		)
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error":   "Invalid or expired token",
			"details": err.Error(),
		})
	}

	// Set token as HTTP-only cookie
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(h.config.Auth.JWTExpirationHours * 3600),
	}
	c.SetCookie(cookie)

	// Fetch user data from database
	user, err := h.getUserByID(c.Request().Context(), claims.UserID)
	if err != nil {
		h.logger.Errorw("Failed to fetch user", "error", err, "user_id", claims.UserID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch user data",
		})
	}

	return c.JSON(http.StatusOK, user)
}

// RefreshToken godoc
// @Summary Refresh JWT token
// @Description Get a new JWT token using an existing valid token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token request"
// @Success 200 {object} CallbackResponse
// @Failure 400 {object} map[string]string
// @Router /auth/refresh [post]
func (h *Handler) RefreshToken(c echo.Context) error {
	var req RefreshRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// Refresh the token
	newToken, err := h.jwtManager.RefreshToken(req.Token)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid or expired token",
		})
	}

	// Validate to get claims
	claims, err := h.jwtManager.ValidateToken(newToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to validate new token",
		})
	}

	// Fetch user data from database
	user, err := h.getUserByID(c.Request().Context(), claims.UserID)
	if err != nil {
		h.logger.Errorw("Failed to fetch user", "error", err, "user_id", claims.UserID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch user data",
		})
	}

	return c.JSON(http.StatusOK, CallbackResponse{
		Token:     newToken,
		ExpiresAt: claims.ExpiresAt.Time,
		User:      *user,
	})
}

// ValidateToken godoc
// @Summary Validate JWT token
// @Description Validate a JWT token and return user info
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} UserInfo
// @Failure 401 {object} map[string]string
// @Router /auth/validate [get]
func (h *Handler) ValidateToken(c echo.Context) error {
	// Token is already validated by middleware
	// Get claims from context (set by middleware)
	claims, ok := c.Get("user").(*auth.JWTClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid token claims",
		})
	}

	// Fetch user data from database
	user, err := h.getUserByID(c.Request().Context(), claims.UserID)
	if err != nil {
		h.logger.Errorw("Failed to fetch user", "error", err, "user_id", claims.UserID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch user data",
		})
	}

	return c.JSON(http.StatusOK, user)
}

// Logout godoc
// @Summary Logout user
// @Description Logout user and clear auth cookie
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string
// @Router /auth/logout [post]
func (h *Handler) Logout(c echo.Context) error {
	// Clear the auth cookie
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1, // Delete the cookie
	}
	c.SetCookie(cookie)

	// In a stateless JWT system, logout is handled client-side
	// For token blacklisting, you would add the token to a blacklist here
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

// findOrCreateUser finds an existing user or creates a new one
func (h *Handler) findOrCreateUser(ctx context.Context, userInfo *oidc.UserInfo, provider string) (*UserInfo, error) {
	// Try to find existing user by OIDC provider and sub
	dbUser, err := h.store.Queries().GetUserByOIDC(ctx, db.GetUserByOIDCParams{
		OidcProvider: provider,
		OidcSub:      userInfo.Sub,
	})
	if err == nil {
		// User found, convert to UserInfo
		return convertDBUserToUserInfo(dbUser.ID, dbUser.Email, dbUser.Name, string(dbUser.Role), dbUser.ClientID), nil
	}

	if err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	// User doesn't exist, create new user
	// Default role: stakeholder (least privileged)
	// Admin should manually assign proper roles
	createdUser, err := h.store.Queries().CreateUser(ctx, db.CreateUserParams{
		Email:        userInfo.Email,
		Name:         userInfo.Name,
		OidcProvider: provider,
		OidcSub:      userInfo.Sub,
		Role:         db.UserRoleEnumStakeholder,
		ClientID:     pgtype.UUID{Valid: false}, // NULL for new users
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	h.logger.Infow("New user created", "user_id", createdUser.ID, "email", createdUser.Email, "provider", provider)

	return convertDBUserToUserInfo(createdUser.ID, createdUser.Email, createdUser.Name, string(createdUser.Role), createdUser.ClientID), nil
}

// getUserByID fetches user data from database by user ID
func (h *Handler) getUserByID(ctx context.Context, userID uuid.UUID) (*UserInfo, error) {
	dbUser, err := h.store.Queries().GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	return convertDBUserToUserInfo(dbUser.ID, dbUser.Email, dbUser.Name, string(dbUser.Role), dbUser.ClientID), nil
}

// updateLastLogin updates the user's last login timestamp
func (h *Handler) updateLastLogin(ctx context.Context, userID uuid.UUID) error {
	return h.store.Queries().UpdateUserLastLogin(ctx, userID)
}

// convertDBUserToUserInfo converts database user types to API UserInfo
func convertDBUserToUserInfo(id uuid.UUID, email, name, role string, clientID pgtype.UUID) *UserInfo {
	user := &UserInfo{
		ID:    id,
		Email: email,
		Name:  name,
		Role:  role,
	}

	if clientID.Valid {
		cid := clientID.Bytes
		parsedUUID, err := uuid.FromBytes(cid[:])
		if err == nil {
			user.ClientID = &parsedUUID
		}
	}

	return user
}
