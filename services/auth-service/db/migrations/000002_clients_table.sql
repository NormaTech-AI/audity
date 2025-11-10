-- Minimal clients table for auth-service queries
-- This is a reference table, actual client management is in tenant-service
CREATE TABLE IF NOT EXISTS clients (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email_domain VARCHAR(255)
);

CREATE INDEX IF NOT EXISTS idx_clients_email_domain ON clients(email_domain);
