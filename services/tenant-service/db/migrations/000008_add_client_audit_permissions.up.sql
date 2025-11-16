-- Add tenant-level permissions for client audit module
-- and assign them to poc_client and stakeholder roles

-- ============================================
-- INSERT PERMISSIONS (idempotent)
-- ============================================
INSERT INTO permissions (name, resource, action, description)
VALUES 
    ('audit:list',  'client_audit', 'list',  'List available audits for the user'),
    ('audit:read',  'client_audit', 'read',  'Read audit details and questions'),
    ('audit:submit','client_audit', 'submit','Submit answers or drafts for audit questions')
ON CONFLICT (name) DO NOTHING;

-- ============================================
-- ENSURE ROLES EXIST (idempotent)
-- ============================================
INSERT INTO roles (name, description)
VALUES
    ('poc_client', 'Client point of contact (external)'),
    ('stakeholder', 'Client stakeholder user')
ON CONFLICT (name) DO NOTHING;

-- ============================================
-- ASSIGN PERMISSIONS TO ROLES (idempotent)
-- ============================================
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name IN ('audit:list','audit:read','audit:submit')
WHERE r.name IN ('poc_client','stakeholder')
ON CONFLICT DO NOTHING;