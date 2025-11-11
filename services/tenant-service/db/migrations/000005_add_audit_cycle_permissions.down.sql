-- Remove role permissions for audit_cycles
DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE resource = 'audit_cycles'
);

-- Remove audit_cycles permissions
DELETE FROM permissions WHERE resource = 'audit_cycles';
