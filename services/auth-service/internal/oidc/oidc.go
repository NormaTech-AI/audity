package oidc

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
)

// OIDCProvider represents an OIDC authentication provider
type OIDCProvider struct {
	config *oauth2.Config
	name   string
}

// UserInfo represents user information from OIDC provider
type UserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// NewGoogleProvider creates a Google OIDC provider
func NewGoogleProvider(clientID, clientSecret, redirectURL string) *OIDCProvider {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &OIDCProvider{
		config: config,
		name:   "google",
	}
}

// NewMicrosoftProvider creates a Microsoft OIDC provider
func NewMicrosoftProvider(clientID, clientSecret, redirectURL, tenantID string) *OIDCProvider {
	// Use tenant-specific endpoint if tenantID is provided, otherwise use "common"
	// Note: "common" only works for multi-tenant apps created before 10/15/2018
	
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"openid",
			"profile",
			"email",
			"User.Read", // Required for Microsoft Graph API /me endpoint
		},
		Endpoint: microsoft.AzureADEndpoint(tenantID),
	}

	return &OIDCProvider{
		config: config,
		name:   "microsoft",
	}
}

// GetAuthURL generates the OAuth2 authorization URL
func (p *OIDCProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges the authorization code for a token
func (p *OIDCProvider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	return token, nil
}

// GetUserInfo retrieves user information using the access token
func (p *OIDCProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	client := p.config.Client(ctx, token)

	var userInfoURL string
	switch p.name {
	case "google":
		userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	case "microsoft":
		userInfoURL = "https://graph.microsoft.com/v1.0/me"
	default:
		return nil, fmt.Errorf("unsupported provider: %s", p.name)
	}

	resp, err := client.Get(userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var userInfo UserInfo
	if p.name == "microsoft" {
		// Microsoft Graph API returns different structure
		var msUser struct {
			ID                string `json:"id"`
			Mail              string `json:"mail"`
			UserPrincipalName string `json:"userPrincipalName"`
			DisplayName       string `json:"displayName"`
			GivenName         string `json:"givenName"`
			Surname           string `json:"surname"`
			Sub               string `json:"sub"`
		}
		if err := json.Unmarshal(body, &msUser); err != nil {
			return nil, fmt.Errorf("failed to unmarshal Microsoft user info: %w", err)
		}

		// Use Sub if available (OIDC standard), otherwise fall back to ID
		sub := msUser.Sub
		if sub == "" {
			sub = msUser.ID
		}

		userInfo = UserInfo{
			Sub:           sub,
			Email:         msUser.Mail,
			EmailVerified: true, // Microsoft emails are verified
			Name:          msUser.DisplayName,
			GivenName:     msUser.GivenName,
			FamilyName:    msUser.Surname,
		}

		// Fallback to UserPrincipalName if Mail is empty
		if userInfo.Email == "" {
			userInfo.Email = msUser.UserPrincipalName
		}
	} else if p.name == "google" {
		// Google OAuth2 userinfo returns 'id' instead of 'sub'
		var googleUser struct {
			ID            string `json:"id"`
			Email         string `json:"email"`
			EmailVerified bool   `json:"email_verified"`
			Name          string `json:"name"`
			GivenName     string `json:"given_name"`
			FamilyName    string `json:"family_name"`
			Picture       string `json:"picture"`
		}
		if err := json.Unmarshal(body, &googleUser); err != nil {
			return nil, fmt.Errorf("failed to unmarshal Google user info: %w", err)
		}

		userInfo = UserInfo{
			Sub:           googleUser.ID,
			Email:         googleUser.Email,
			EmailVerified: googleUser.EmailVerified,
			Name:          googleUser.Name,
			GivenName:     googleUser.GivenName,
			FamilyName:    googleUser.FamilyName,
			Picture:       googleUser.Picture,
		}
	} else {
		if err := json.Unmarshal(body, &userInfo); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user info: %w", err)
		}
	}

	return &userInfo, nil
}

// GetProviderName returns the provider name
func (p *OIDCProvider) GetProviderName() string {
	return p.name
}

// GenerateState generates a random state string for OAuth2
func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
