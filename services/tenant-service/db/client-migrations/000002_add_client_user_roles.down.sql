-- Drop client-specific user roles and permissions

DROP TABLE IF EXISTS client_user_roles;
DROP TABLE IF EXISTS client_role_permissions;
DROP TABLE IF EXISTS client_permissions;
DROP TABLE IF EXISTS client_roles;
DROP TABLE IF EXISTS client_users;

DROP TYPE IF EXISTS client_user_role_enum;
