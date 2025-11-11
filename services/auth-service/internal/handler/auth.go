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
	ID             uuid.UUID  `json:"id"`
	Email          string     `json:"email"`
	Name           string     `json:"name"`
	Designation    string     `json:"designation"`     // Job title (e.g., nishaj_admin, auditor, etc.)
	Roles          []string   `json:"roles"`           // RBAC roles from user_roles table
	ClientID       *uuid.UUID `json:"client_id,omitempty"`
	VisibleModules []string   `json:"visible_modules"`
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
		result := convertDBUserToUserInfo(dbUser.ID, dbUser.Email, dbUser.Name, string(dbUser.Designation), dbUser.ClientID)
		// Fetch user roles from user_roles table
		roles, _ := h.getUserRoles(ctx, dbUser.ID)
		result.Roles = roles
		result.VisibleModules = getVisibleModules(roles, string(dbUser.Designation))
		return result, nil
	}

	if err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	// User doesn't exist, create new user
	// Try to auto-assign client based on email domain
	var clientID pgtype.UUID
	emailDomain := extractEmailDomain(userInfo.Email)
	clientDetails := db.Client{}
	if emailDomain != "" {
		client, err := h.store.Queries().GetClientByEmailDomain(ctx, &emailDomain)
		clientDetails = client
		if err == nil {
			// Found matching client, assign it
			clientID = pgtype.UUID{
				Bytes: client.ID,
				Valid: true,
			}
			h.logger.Infow("Auto-assigning client based on email domain",
				"email", userInfo.Email,
				"domain", emailDomain,
				"client_id", client.ID,
				"client_name", client.Name)
		} else if err != pgx.ErrNoRows {
			// Log error but don't fail user creation
			h.logger.Warnw("Failed to query client by email domain",
				"error", err,
				"email", userInfo.Email,
				"domain", emailDomain)
		}
	}
	
	// Default designation: stakeholder (least privileged)
	// Admin should manually assign proper roles via user_roles table
	createdUser, err := h.store.Queries().CreateUser(ctx, db.CreateUserParams{
		Email:        userInfo.Email,
		Name:         userInfo.Name,
		OidcProvider: provider,
		OidcSub:      userInfo.Sub,
		Designation:  db.UserDesignationEnumStakeholder,
		ClientID:     clientID,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Check if the user is poc of the client, yes them assign poc_client role
	if clientDetails.PocEmail != nil {
		if *clientDetails.PocEmail == userInfo.Email {
			h.store.Queries().CreateUserRole(ctx, db.CreateUserRoleParams{
				UserID:   createdUser.ID,
				RoleID:   uuid.MustParse("55555555-5555-5555-5555-555555555555"),
				ClientID: clientID,
			})
		}else{
			h.store.Queries().CreateUserRole(ctx, db.CreateUserRoleParams{
				UserID:   createdUser.ID,
				RoleID:   uuid.MustParse("66666666-6666-6666-6666-666666666666"),
				ClientID: clientID,
			})
		}
	}

	if clientID.Valid {
		h.logger.Infow("New user created with auto-assigned client",
			"user_id", createdUser.ID,
			"email", createdUser.Email,
			"provider", provider,
			"client_id", clientID.Bytes)
	} else {
		h.logger.Infow("New user created without client assignment",
			"user_id", createdUser.ID,
			"email", createdUser.Email,
			"provider", provider)
	}

	result := convertDBUserToUserInfo(createdUser.ID, createdUser.Email, createdUser.Name, string(createdUser.Designation), createdUser.ClientID)
	// Fetch user roles from user_roles table (new user might not have roles yet)
	roles, _ := h.getUserRoles(ctx, createdUser.ID)
	result.Roles = roles
	result.VisibleModules = getVisibleModules(roles, string(createdUser.Designation))
	return result, nil
}

// getUserByID fetches user data from database by user ID
func (h *Handler) getUserByID(ctx context.Context, userID uuid.UUID) (*UserInfo, error) {
	dbUser, err := h.store.Queries().GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	result := convertDBUserToUserInfo(dbUser.ID, dbUser.Email, dbUser.Name, string(dbUser.Designation), dbUser.ClientID)
	// Fetch user roles from user_roles table
	roles, _ := h.getUserRoles(ctx, dbUser.ID)
	result.Roles = roles
	result.VisibleModules = getVisibleModules(roles, string(dbUser.Designation))
	h.logger.Infow("User fetched successfully", "user_id", dbUser.ID, "email", dbUser.Email, "roles", roles)
	return result, nil
}

// updateLastLogin updates the user's last login timestamp
func (h *Handler) updateLastLogin(ctx context.Context, userID uuid.UUID) error {
	return h.store.Queries().UpdateUserLastLogin(ctx, userID)
}

// convertDBUserToUserInfo converts database user types to API UserInfo
func convertDBUserToUserInfo(id uuid.UUID, email, name, designation string, clientID pgtype.UUID) *UserInfo {
	user := &UserInfo{
		ID:             id,
		Email:          email,
		Name:           name,
		Designation:    designation,
		Roles:          []string{}, // Will be populated by caller
		VisibleModules: []string{}, // Will be populated based on roles
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

// getUserRoles fetches user roles from the user_roles table
func (h *Handler) getUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	// Query to get user roles from user_roles table
	query := `
		SELECT r.name 
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.name
	`
	
	rows, err := h.store.GetPool().Query(ctx, query, userID)
	if err != nil {
		return []string{}, nil // Return empty array if no roles found
	}
	defer rows.Close()
	
	roles := []string{}
	for rows.Next() {
		var roleName string
		if err := rows.Scan(&roleName); err != nil {
			continue
		}
		roles = append(roles, roleName)
	}
	
	return roles, nil
}

// getVisibleModules returns the list of modules visible to a user based on their roles
// If user has admin role, they see all modules
// Otherwise, modules are determined by their designation as a fallback
func getVisibleModules(roles []string, designation string) []string {
	// Check if user has admin role
	for _, role := range roles {
		switch role {
			case "admin":
				return []string{"Dashboard", "Clients", "Users", "Roles & Permissions", "Assessments", "Frameworks", "Audit Cycles"}
			case "nishaj_admin":
				return []string{"Dashboard", "Clients", "Users", "Roles & Permissions", "Assessments", "Frameworks", "Audit Cycles"}
			case "auditor":
				return []string{"Dashboard", "Clients", "Assessments"}
			case "team_member":
				return []string{"Dashboard", "Assessments"}
			case "poc_internal", "poc_client":
				return []string{"Dashboard", "Audit", "Roles & Permissions"}
			case "stakeholder":
				return []string{"Dashboard", "Assessments"}
			default:
				return []string{"Dashboard"}
		}
	}
	
	// Default for users with roles but not admin
	return []string{"Dashboard"}
}

// extractEmailDomain extracts the domain from an email address
// Example: "user@example.com" -> "example.com"
func extractEmailDomain(email string) string {
	parts := splitEmail(email)
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// splitEmail splits an email address into local and domain parts
func splitEmail(email string) []string {
	for i := len(email) - 1; i >= 0; i-- {
		if email[i] == '@' {
			return []string{email[:i], email[i+1:]}
		}
	}
	return []string{email}
}
