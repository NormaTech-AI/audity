package microsoftmail

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"go.uber.org/zap"
)

// --- Configuration ---

// Config holds the Microsoft Graph API credentials and configuration.
type Config struct {
	TenantID     string
	ClientID     string
	ClientSecret string
	SenderEmail  string
}

// Client represents a Microsoft Graph API client for sending emails.
type Client struct {
	config      Config
	accessToken string
}

// NewClient creates a new Microsoft Graph API client with the given configuration.
func NewClient(config Config, log *zap.SugaredLogger) (*Client, error) {
	return &Client{
		config: config,
	}, nil
}

// --- Structs for Microsoft Graph API ---

// TokenResponse is used to decode the OAuth access token response.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

// Below structs are for building the sendMail JSON payload.
type EmailAddress struct {
	Address string `json:"address"`
}

type Recipient struct {
	EmailAddress EmailAddress `json:"emailAddress"`
}

type Body struct {
	ContentType string `json:"contentType"` // "Text" or "HTML"
	Content     string `json:"content"`
}

type Message struct {
	Subject      string      `json:"subject"`
	Body         Body        `json:"body"`
	ToRecipients []Recipient `json:"toRecipients"`
}

type SendMailPayload struct {
	Message Message `json:"message"`
}

// getAccessToken fetches an OAuth 2.0 access token from Microsoft.
func (c *Client) getAccessToken(logger *zap.SugaredLogger) error {
	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", c.config.TenantID)
	
	logger.Infow("Requesting access token", 
		"tokenURL", tokenURL,
		"tenantID", c.config.TenantID,
		"clientID", c.config.ClientID,
		"senderEmail", c.config.SenderEmail)

	// This data is sent as application/x-www-form-urlencoded
	data := url.Values{}
	data.Set("client_id", c.config.ClientID)
	data.Set("client_secret", c.config.ClientSecret)
	data.Set("grant_type", "client_credentials")
	data.Set("scope", "https://graph.microsoft.com/.default")

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		logger.Errorw("error creating token request", "error", err)
		return fmt.Errorf("error creating token request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Errorw("error performing token request", "error", err)
		return fmt.Errorf("error performing token request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read the body for detailed error message
		var errorBody bytes.Buffer
		errorBody.ReadFrom(resp.Body)
		logger.Errorw("failed to get token", 
			"status", resp.Status,
			"statusCode", resp.StatusCode,
			"body", errorBody.String(),
			"tenantID", c.config.TenantID)
		return fmt.Errorf("failed to get token, status: %s, body: %s", resp.Status, errorBody.String())
	}

	var tokenResponse TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		logger.Errorw("error decoding token response", "error", err)
		return fmt.Errorf("error decoding token response: %v", err)
	}

	c.accessToken = tokenResponse.AccessToken
	logger.Info("Access token acquired successfully")
	return nil
}

// EmailParams contains the parameters for sending an email.
type EmailParams struct {
	To          []string // List of recipient email addresses
	Subject     string
	Body        string
	ContentType string // "HTML" or "Text", defaults to "HTML"
}

// sendMail uses the Graph API to send an email.
func (c *Client) sendMail(params EmailParams, logger *zap.SugaredLogger) error {
	// This is the API endpoint. We send mail "from" the user specified.
	graphURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/sendMail", c.config.SenderEmail)

	// Set default content type if not specified
	contentType := params.ContentType
	if contentType == "" {
		contentType = "HTML"
	}

	// Build recipients list
	recipients := make([]Recipient, len(params.To))
	for i, email := range params.To {
		recipients[i] = Recipient{
			EmailAddress: EmailAddress{
				Address: email,
			},
		}
	}

	// 1. Construct the email payload
	payload := SendMailPayload{
		Message: Message{
			Subject:      params.Subject,
			Body:         Body{
				ContentType: contentType,
				Content:     params.Body,
			},
			ToRecipients: recipients,
		},
	}

	// 2. Marshal payload to JSON
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshalling json: %v", err)
	}

	// 3. Create the HTTP request
	req, err := http.NewRequest("POST", graphURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		logger.Errorw("error creating graph request", "error", err)
		return fmt.Errorf("error creating graph request: %v", err)
	}

	// 4. Set headers
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	// 5. Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Errorw("error performing graph request", "error", err)
		return fmt.Errorf("error performing graph request: %v", err)
	}
	defer resp.Body.Close()

	// The 'sendMail' endpoint returns 202 Accepted on success
	if resp.StatusCode != http.StatusAccepted {
		// Read the body for a more detailed error message from Graph
		var errorBody bytes.Buffer
		errorBody.ReadFrom(resp.Body)
		logger.Errorw("failed to send mail", "status", resp.Status, "body", errorBody.String())
		return fmt.Errorf("failed to send mail, status: %s, body: %s", resp.Status, errorBody.String())
	}
	logger.Info("email sent successfully")
	return nil
}

// SendEmail sends an email using the Microsoft Graph API.
// It automatically handles token acquisition if needed.
func (c *Client) SendEmail(params EmailParams, logger *zap.SugaredLogger) error {
	// Get access token if we don't have one
	logger.Info("Sending email...", "params", params)
	logger.Info(c)
	if c.accessToken == "" {
		if err := c.getAccessToken(logger); err != nil {
			logger.Errorw("failed to get access token", "error", err)
			return fmt.Errorf("failed to get access token: %w", err)
		}
	}

	// Send the email
	if err := c.sendMail(params, logger); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}