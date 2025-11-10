-- Add client-specific user roles and permissions to each client database
-- This allows each client to have their own RBAC system

-- ============================================
-- ENUMS
-- ============================================

-- User role enum for client-specific roles
CREATE TYPE client_user_role_enum AS ENUM (
    'client_admin',      -- Client administrator (full access to client data)
    'poc',               -- Point of Contact (can manage audits and delegate)
    'stakeholder',       -- Stakeholder (can answer assigned questions)
    'viewer'             -- Read-only access
);

-- ============================================
-- TABLES
-- ============================================

-- Client Users
-- Maps tenant_db users to this client with client-specific roles
CREATE TABLE client_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_user_id UUID NOT NULL UNIQUE, -- References users.id in tenant_db
    email VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role client_user_role_enum NOT NULL DEFAULT 'viewer',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_login TIMESTAMP WITH TIME ZONE
);

-- Roles table (client-specific roles)
CREATE TABLE client_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Permissions table (client-specific permissions)
CREATE TABLE client_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    resource VARCHAR(100) NOT NULL, -- e.g., 'audits', 'submissions', 'reports'
    action VARCHAR(50) NOT NULL, -- e.g., 'create', 'read', 'update', 'delete'
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Role permissions mapping
CREATE TABLE client_role_permissions (
    role_id UUID NOT NULL REFERENCES client_roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES client_permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- User roles mapping (users can have multiple roles)
CREATE TABLE client_user_roles (
    user_id UUID NOT NULL REFERENCES client_users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES client_roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- ============================================
-- INDEXES
-- ============================================

CREATE INDEX idx_client_users_tenant_user_id ON client_users(tenant_user_id);
CREATE INDEX idx_client_users_email ON client_users(email);
CREATE INDEX idx_client_users_role ON client_users(role);
CREATE INDEX idx_client_users_is_active ON client_users(is_active);
CREATE INDEX idx_client_permissions_resource ON client_permissions(resource);

-- ============================================
-- TRIGGERS
-- ============================================

CREATE TRIGGER update_client_users_updated_at BEFORE UPDATE ON client_users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- COMMENTS
-- ============================================

COMMENT ON TABLE client_users IS 'Client-specific user mappings with roles';
COMMENT ON TABLE client_roles IS 'Client-specific RBAC roles';
COMMENT ON TABLE client_permissions IS 'Client-specific RBAC permissions';
COMMENT ON TABLE client_role_permissions IS 'Mapping between roles and permissions';
COMMENT ON TABLE client_user_roles IS 'Mapping between users and roles';
