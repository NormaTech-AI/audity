-- Insert permissions for audit_cycles resource
INSERT INTO permissions (name, resource, action, description) VALUES
    ('audit_cycles:create', 'audit_cycles', 'create', 'Create audit cycles'),
    ('audit_cycles:read', 'audit_cycles', 'read', 'View audit cycle details'),
    ('audit_cycles:update', 'audit_cycles', 'update', 'Update audit cycles'),
    ('audit_cycles:delete', 'audit_cycles', 'delete', 'Delete audit cycles'),
    ('audit_cycles:list', 'audit_cycles', 'list', 'List all audit cycles'),
    ('audit_cycles:manage_clients', 'audit_cycles', 'manage_clients', 'Add/remove clients from audit cycles'),
    ('audit_cycles:assign_frameworks', 'audit_cycles', 'assign_frameworks', 'Assign frameworks to clients in audit cycles');

-- Assign all audit cycle permissions to nishaj_admin role (full access)
INSERT INTO role_permissions (role_id, permission_id)
SELECT '11111111-1111-1111-1111-111111111111', id FROM permissions
WHERE resource = 'audit_cycles';

-- Assign audit cycle permissions to auditor role (read, list, review)
INSERT INTO role_permissions (role_id, permission_id)
SELECT '22222222-2222-2222-2222-222222222222', id FROM permissions
WHERE name IN (
    'audit_cycles:read',
    'audit_cycles:list'
);

-- Assign audit cycle permissions to team_member role (read only)
INSERT INTO role_permissions (role_id, permission_id)
SELECT '33333333-3333-3333-3333-333333333333', id FROM permissions
WHERE name IN (
    'audit_cycles:read',
    'audit_cycles:list'
);

-- Assign audit cycle permissions to poc_internal role (read, list)
INSERT INTO role_permissions (role_id, permission_id)
SELECT '44444444-4444-4444-4444-444444444444', id FROM permissions
WHERE name IN (
    'audit_cycles:read',
    'audit_cycles:list'
);

-- poc_client and stakeholder roles do not get audit cycle permissions
-- as they should not see the audit cycle management interface
