# Microsoft Mail Package

A reusable Go package for sending emails via Microsoft Graph API.

## Features

- OAuth 2.0 authentication with Microsoft Graph API
- Send emails with HTML or plain text content
- Support for multiple recipients
- Automatic token management
- Clean, reusable API

## Installation

```bash
go get github.com/NormaTech-AI/audity/packages/go/microsoft-mail
```

## Prerequisites

You need to set up an Azure AD application with the following:

1. **Azure AD App Registration** with these permissions:
   - `Mail.Send` (Application permission)
   
2. **Required credentials**:
   - Tenant ID
   - Client ID
   - Client Secret
   - Sender Email (the email address that will send emails)

## Usage

### Basic Example

```go
package main

import (
    "log"
    microsoftmail "github.com/NormaTech-AI/audity/packages/go/microsoft-mail"
)

func main() {
    // Create configuration
    config := microsoftmail.Config{
        TenantID:     "your-tenant-id",
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
        SenderEmail:  "sender@yourdomain.com",
    }

    // Create client
    client := microsoftmail.NewClient(config)

    // Send email
    err := client.SendEmail(microsoftmail.EmailParams{
        To:          []string{"recipient@example.com"},
        Subject:     "Hello from Microsoft Graph API",
        Body:        "<h1>Test Email</h1><p>This is a test email.</p>",
        ContentType: "HTML", // or "Text"
    })

    if err != nil {
        log.Fatalf("Failed to send email: %v", err)
    }

    log.Println("Email sent successfully!")
}
```

### Sending to Multiple Recipients

```go
err := client.SendEmail(microsoftmail.EmailParams{
    To: []string{
        "recipient1@example.com",
        "recipient2@example.com",
        "recipient3@example.com",
    },
    Subject:     "Team Notification",
    Body:        "<p>This email goes to multiple recipients.</p>",
    ContentType: "HTML",
})
```

### Sending Plain Text Email

```go
err := client.SendEmail(microsoftmail.EmailParams{
    To:          []string{"recipient@example.com"},
    Subject:     "Plain Text Email",
    Body:        "This is a plain text email.",
    ContentType: "Text",
})
```

### Using Environment Variables

```go
import (
    "os"
    microsoftmail "github.com/NormaTech-AI/audity/packages/go/microsoft-mail"
)

config := microsoftmail.Config{
    TenantID:     os.Getenv("MICROSOFT_TENANT_ID"),
    ClientID:     os.Getenv("MICROSOFT_CLIENT_ID"),
    ClientSecret: os.Getenv("MICROSOFT_CLIENT_SECRET"),
    SenderEmail:  os.Getenv("MICROSOFT_SENDER_EMAIL"),
}

client := microsoftmail.NewClient(config)
```

## API Reference

### Types

#### `Config`
Configuration for Microsoft Graph API authentication.

```go
type Config struct {
    TenantID     string // Azure AD Tenant ID
    ClientID     string // Azure AD Application (client) ID
    ClientSecret string // Azure AD Client Secret
    SenderEmail  string // Email address to send from
}
```

#### `EmailParams`
Parameters for sending an email.

```go
type EmailParams struct {
    To          []string // List of recipient email addresses
    Subject     string   // Email subject
    Body        string   // Email body content
    ContentType string   // "HTML" or "Text" (defaults to "HTML")
}
```

### Functions

#### `NewClient(config Config) *Client`
Creates a new Microsoft Graph API client.

#### `(*Client) SendEmail(params EmailParams) error`
Sends an email using the Microsoft Graph API. Automatically handles token acquisition.

## Error Handling

The package returns descriptive errors for common issues:

```go
err := client.SendEmail(params)
if err != nil {
    // Handle error - could be authentication, network, or API errors
    log.Printf("Error sending email: %v", err)
}
```

## License

Part of the Audity project.
