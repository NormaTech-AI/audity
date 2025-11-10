-- Create enum types
-- Note: This represents user designation/job title, not RBAC roles
-- RBAC roles are managed in the user_roles table in tenant-service
CREATE TYPE user_designation_enum AS ENUM ('nishaj_admin', 'auditor', 'team_member', 'poc_internal', 'poc_client', 'stakeholder');

-- Users table (all users across tenants - for OIDC mapping)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    oidc_provider VARCHAR(50) NOT NULL, -- 'google' or 'microsoft'
    oidc_sub VARCHAR(255) NOT NULL, -- Subject from OIDC provider
    designation user_designation_enum NOT NULL, -- User's job title/position
    client_id UUID, -- NULL for Nishaj internal users
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login TIMESTAMP WITH TIME ZONE,
    UNIQUE(oidc_provider, oidc_sub)
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_client_id ON users(client_id);
CREATE INDEX idx_users_designation ON users(designation);
CREATE INDEX idx_users_oidc ON users(oidc_provider, oidc_sub);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
