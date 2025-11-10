# Automatic Client Assignment

## Overview

The auth service now automatically assigns users to clients based on their email domain during the OIDC authentication flow.

## How It Works

1. **User Login**: When a user logs in via Google or Microsoft OIDC
2. **Email Domain Extraction**: The system extracts the domain from the user's email
   - Example: `john@example.com` → domain is `example.com`
3. **Client Lookup**: The system queries the `clients` table for a matching `email_domain`
4. **Auto-Assignment**: If a match is found, the user's `client_id` is automatically set
5. **User Creation**: The user is created with the assigned `client_id`

## Database Setup

### Required Migration

The `clients` table must have the `email_domain` column:

```sql
ALTER TABLE clients ADD COLUMN email_domain VARCHAR(255);
CREATE INDEX idx_clients_email_domain ON clients(email_domain);
```

This is already included in migration `000003_add_client_email_domain.up.sql` in the tenant-service.

### Setting Email Domains

To enable auto-assignment for a client, set their `email_domain`:

```sql
UPDATE clients 
SET email_domain = 'example.com' 
WHERE name = 'Example Corp';
```

## Behavior

### New Users
- **With matching domain**: Automatically assigned to the client with `role = 'stakeholder'`
- **Without matching domain**: Created with `client_id = NULL` and `role = 'stakeholder'`

### Existing Users
- Existing users are NOT reassigned
- Auto-assignment only happens during initial user creation

### Multiple Clients
- If multiple clients have the same `email_domain`, the first match is used
- It's recommended to keep `email_domain` unique per client

## Logging

The system logs auto-assignment events:

```json
{
  "level": "info",
  "msg": "Auto-assigning client based on email domain",
  "email": "john@example.com",
  "domain": "example.com",
  "client_id": "...",
  "client_name": "Example Corp"
}
```

## Security Considerations

1. **Email Verification**: Both Google and Microsoft verify email addresses before issuing tokens
2. **Domain Ownership**: Only set `email_domain` for domains you control or trust
3. **Role Assignment**: All auto-assigned users start with the least privileged role (`stakeholder`)
4. **Manual Override**: Admins can manually change user assignments and roles after creation

## API Changes

No API changes are required. The auto-assignment happens transparently during the OIDC callback flow.

## Testing

To test auto-assignment:

1. Create a client with an email domain:
   ```sql
   INSERT INTO clients (name, poc_email, status, email_domain)
   VALUES ('Test Corp', 'admin@testcorp.com', 'active', 'testcorp.com');
   ```

2. Log in with a user having email `user@testcorp.com`

3. Check the user's `client_id` in the database:
   ```sql
   SELECT id, email, client_id FROM users WHERE email = 'user@testcorp.com';
   ```

## Troubleshooting

### User not assigned despite matching domain

1. Check if the `email_domain` is set correctly in the `clients` table
2. Check the auth-service logs for any errors
3. Verify the migration has been applied to the database

### Multiple assignments

If you need to change a user's client assignment, update it manually:

```sql
UPDATE users 
SET client_id = '<new-client-uuid>' 
WHERE email = 'user@example.com';
```
