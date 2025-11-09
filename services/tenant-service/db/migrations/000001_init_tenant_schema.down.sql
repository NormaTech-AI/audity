-- Drop triggers
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_client_frameworks_updated_at ON client_frameworks;
DROP TRIGGER IF EXISTS update_compliance_frameworks_updated_at ON compliance_frameworks;
DROP TRIGGER IF EXISTS update_clients_updated_at ON clients;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse order
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS client_frameworks;
DROP TABLE IF EXISTS compliance_frameworks;
DROP TABLE IF EXISTS client_buckets;
DROP TABLE IF EXISTS client_databases;
DROP TABLE IF EXISTS clients;

-- Drop enum types
DROP TYPE IF EXISTS audit_status_enum;
DROP TYPE IF EXISTS user_role_enum;
DROP TYPE IF EXISTS client_status_enum;
