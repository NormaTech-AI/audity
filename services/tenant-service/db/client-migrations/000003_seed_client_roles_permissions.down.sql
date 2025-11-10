-- Remove seeded client roles and permissions

DELETE FROM client_role_permissions;
DELETE FROM client_permissions;
DELETE FROM client_roles;
