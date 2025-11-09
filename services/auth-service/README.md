# Auth Service

Authentication microservice for the TPRM Audit Platform. Handles OAuth2/OIDC authentication with Google and Microsoft, JWT token generation and validation.

## Features

- ✅ OAuth2/OIDC Integration (Google & Microsoft)
- ✅ JWT Token Generation & Validation
- ✅ Token Refresh
- ✅ User Auto-Registration
- ✅ Stateless Authentication
- ✅ Swagger API Documentation

## Architecture

```
┌──────────────┐
│   Frontend   │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Auth Service │ (Port 8082)
│  - OAuth2    │
│  - JWT       │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│  tenant_db   │
│ (PostgreSQL) │
└──────────────┘
```

## API Endpoints

### Public Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Service info |
| GET | `/health` | Health check |
| GET | `/auth/login/:provider` | Initiate OAuth login (google/microsoft) |
| GET | `/auth/callback` | OAuth callback handler |
| POST | `/auth/refresh` | Refresh JWT token |
| POST | `/auth/logout` | Logout (client-side) |

### Protected Endpoints (Require JWT)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/auth/validate` | Validate JWT token |

### Swagger Documentation

Access API docs at: `http://localhost:8082/swagger/index.html`

## Configuration

### Environment Variables

All config can be set via environment variables with `AUDITY_AUTH_` prefix:

```bash
# Server
AUDITY_AUTH_SERVER_PORT=8082
AUDITY_AUTH_SERVER_HOST=0.0.0.0
AUDITY_AUTH_SERVER_ENV=development

# Database
AUDITY_AUTH_DATABASE_TENANT_DB_URL=postgres://root:password@localhost:5432/tenant_db?sslmode=disable

# JWT
AUDITY_AUTH_AUTH_JWT_SECRET=your-super-secret-jwt-key-change-in-production-must-be-32-chars
AUDITY_AUTH_AUTH_JWT_EXPIRATION_HOURS=24

# OAuth Providers
AUDITY_AUTH_AUTH_GOOGLE_CLIENT_ID=your-google-client-id
AUDITY_AUTH_AUTH_GOOGLE_CLIENT_SECRET=your-google-client-secret
AUDITY_AUTH_AUTH_MICROSOFT_CLIENT_ID=your-microsoft-client-id
AUDITY_AUTH_AUTH_MICROSOFT_CLIENT_SECRET=your-microsoft-client-secret

# URLs
AUDITY_AUTH_AUTH_REDIRECT_URL=http://localhost:8082/auth/callback
AUDITY_AUTH_AUTH_FRONTEND_URL=http://localhost:5173
```

### OAuth Provider Setup

#### Google OAuth2

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing
3. Enable Google+ API
4. Create OAuth 2.0 credentials
5. Add authorized redirect URI: `http://localhost:8082/auth/callback`
6. Copy Client ID and Client Secret to config

#### Microsoft OAuth2

1. Go to [Azure Portal](https://portal.azure.com/)
2. Navigate to Azure Active Directory > App registrations
3. Create new registration
4. Add redirect URI: `http://localhost:8082/auth/callback`
5. Create client secret
6. Copy Application (client) ID and secret to config

## Authentication Flow

### 1. Login Initiation

```
GET /auth/login/google
```

Response:
```json
{
  "auth_url": "https://accounts.google.com/o/oauth2/v2/auth?...",
  "provider": "google"
}
```

Frontend redirects user to `auth_url`.

### 2. OAuth Callback

After user authorizes, provider redirects to:
```
GET /auth/callback?code=...&state=...
```

Service:
1. Exchanges code for access token
2. Fetches user info from provider
3. Creates/updates user in database
4. Generates JWT token
5. Redirects to frontend with token

### 3. Using JWT Token

Include token in Authorization header:
```
Authorization: Bearer <jwt_token>
```

### 4. Token Validation

Other services can validate tokens by calling:
```
GET /auth/validate
Authorization: Bearer <jwt_token>
```

Response:
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "name": "John Doe",
  "role": "stakeholder",
  "client_id": "uuid"
}
```

### 5. Token Refresh

```
POST /auth/refresh
{
  "token": "old_jwt_token"
}
```

Response:
```json
{
  "token": "new_jwt_token",
  "expires_at": "2025-11-07T12:00:00Z",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "stakeholder"
  }
}
```

## JWT Token Structure

```json
{
  "user_id": "uuid",
  "email": "user@example.com",
  "name": "John Doe",
  "role": "stakeholder",
  "client_id": "uuid",
  "exp": 1730976000,
  "iat": 1730889600,
  "iss": "audity-auth-service",
  "sub": "uuid"
}
```

## User Roles

Default role for new users: `stakeholder` (least privileged)

Available roles:
- `nishaj_admin` - Full system access
- `auditor` - Review and approve submissions
- `team_member` - Support staff
- `poc_internal` - Internal relationship manager
- `poc_client` - Client point of contact
- `stakeholder` - Client employee (default)

Admins must manually assign proper roles after user registration.

## Development

### Prerequisites

```bash
# Install Go tools
make install-tools
```

### Run Locally

```bash
# Development mode (hot reload)
make dev

# Or build and run
make build
./bin/auth-service
```

### Generate Swagger Docs

```bash
make swagger
```

### Run Tests

```bash
make test
```

## Docker

### Build

```bash
docker build -t auth-service .
```

### Run

```bash
docker run -p 8082:8082 \
  -e AUDITY_AUTH_DATABASE_TENANT_DB_URL=postgres://... \
  -e AUDITY_AUTH_AUTH_JWT_SECRET=... \
  auth-service
```

## Security Considerations

### Production Checklist

- [ ] Change JWT secret to strong random value (32+ characters)
- [ ] Use HTTPS for all endpoints
- [ ] Configure OAuth redirect URLs for production domain
- [ ] Enable CORS only for trusted origins
- [ ] Implement rate limiting
- [ ] Add token blacklisting for logout (use Redis)
- [ ] Store OAuth state in Redis instead of memory
- [ ] Enable database connection encryption
- [ ] Set up proper logging and monitoring
- [ ] Implement token rotation
- [ ] Add IP whitelisting for sensitive endpoints

### JWT Best Practices

- Tokens expire after 24 hours (configurable)
- Use refresh tokens for long-lived sessions
- Validate tokens on every request
- Never store sensitive data in JWT payload
- Use HTTPS to prevent token interception

## Troubleshooting

### OAuth Provider Errors

**Error: redirect_uri_mismatch**
- Ensure redirect URI in provider settings matches exactly: `http://localhost:8082/auth/callback`

**Error: invalid_client**
- Verify client ID and secret are correct
- Check if OAuth app is enabled

### Database Connection Issues

**Error: failed to connect to database**
- Verify PostgreSQL is running
- Check connection string format
- Ensure database `tenant_db` exists

### Token Validation Failures

**Error: invalid or expired token**
- Token may have expired (default 24 hours)
- Use refresh endpoint to get new token
- Verify JWT secret matches between services

## Inter-Service Communication

Other microservices can validate tokens by:

1. **Direct validation** (recommended):
   ```go
   jwtManager := auth.NewJWTManager(jwtSecret, expirationHours)
   claims, err := jwtManager.ValidateToken(tokenString)
   ```

2. **HTTP validation**:
   ```bash
   curl -H "Authorization: Bearer <token>" http://auth-service:8082/auth/validate
   ```

## Monitoring

### Health Check

```bash
curl http://localhost:8082/health
```

Response:
```json
{
  "status": "ok",
  "database": "connected",
  "service": "auth-service"
}
```

### Metrics (Future)

- Total logins by provider
- Active sessions
- Token validation rate
- Failed authentication attempts

## Future Enhancements

- [ ] Redis for state storage and token blacklist
- [ ] Multi-factor authentication (MFA)
- [ ] Social login (GitHub, LinkedIn)
- [ ] SAML support for enterprise SSO
- [ ] Session management dashboard
- [ ] Audit logs for authentication events
- [ ] Rate limiting per user/IP
- [ ] Passwordless authentication
- [ ] Biometric authentication support

## License

Proprietary - Nishaj Infotech
