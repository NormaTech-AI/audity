-- Seed default client-specific roles and permissions

-- ============================================
-- INSERT ROLES
-- ============================================

INSERT INTO client_roles (id, name, description) VALUES
    ('11111111-1111-1111-1111-111111111111', 'client_admin', 'Full administrative access to client data'),
    ('22222222-2222-2222-2222-222222222222', 'poc', 'Point of Contact - can manage audits and delegate tasks'),
    ('33333333-3333-3333-3333-333333333333', 'stakeholder', 'Can answer assigned questions and submit evidence'),
    ('44444444-4444-4444-4444-444444444444', 'viewer', 'Read-only access to audit data');

-- ============================================
-- INSERT PERMISSIONS
-- ============================================

-- Audit permissions
INSERT INTO client_permissions (name, resource, action, description) VALUES
    ('audits:read', 'audits', 'read', 'View audit details'),
    ('audits:update', 'audits', 'update', 'Update audit information'),
    ('audits:manage', 'audits', 'manage', 'Full audit management');

-- Question permissions
INSERT INTO client_permissions (name, resource, action, description) VALUES
    ('questions:read', 'questions', 'read', 'View questions'),
    ('questions:assign', 'questions', 'assign', 'Assign questions to stakeholders');

-- Submission permissions
INSERT INTO client_permissions (name, resource, action, description) VALUES
    ('submissions:read', 'submissions', 'read', 'View submissions'),
    ('submissions:create', 'submissions', 'create', 'Create and submit answers'),
    ('submissions:update', 'submissions', 'update', 'Update own submissions'),
    ('submissions:manage', 'submissions', 'manage', 'Manage all submissions');

-- Evidence permissions
INSERT INTO client_permissions (name, resource, action, description) VALUES
    ('evidence:read', 'evidence', 'read', 'View evidence files'),
    ('evidence:upload', 'evidence', 'upload', 'Upload evidence files'),
    ('evidence:delete', 'evidence', 'delete', 'Delete evidence files');

-- Report permissions
INSERT INTO client_permissions (name, resource, action, description) VALUES
    ('reports:read', 'reports', 'read', 'View reports'),
    ('reports:download', 'reports', 'download', 'Download reports');

-- Comment permissions
INSERT INTO client_permissions (name, resource, action, description) VALUES
    ('comments:read', 'comments', 'read', 'View comments'),
    ('comments:create', 'comments', 'create', 'Create comments');

-- User management permissions
INSERT INTO client_permissions (name, resource, action, description) VALUES
    ('users:read', 'users', 'read', 'View client users'),
    ('users:manage', 'users', 'manage', 'Manage client users and roles');

-- ============================================
-- ASSIGN PERMISSIONS TO ROLES
-- ============================================

-- Client Admin: Full access
INSERT INTO client_role_permissions (role_id, permission_id)
SELECT '11111111-1111-1111-1111-111111111111', id FROM client_permissions;

-- POC: Can manage audits, assign questions, view all submissions
INSERT INTO client_role_permissions (role_id, permission_id)
SELECT '22222222-2222-2222-2222-222222222222', id FROM client_permissions
WHERE name IN (
    'audits:read', 'audits:update', 'audits:manage',
    'questions:read', 'questions:assign',
    'submissions:read', 'submissions:create', 'submissions:update', 'submissions:manage',
    'evidence:read', 'evidence:upload', 'evidence:delete',
    'reports:read', 'reports:download',
    'comments:read', 'comments:create',
    'users:read'
);

-- Stakeholder: Can answer assigned questions and submit evidence
INSERT INTO client_role_permissions (role_id, permission_id)
SELECT '33333333-3333-3333-3333-333333333333', id FROM client_permissions
WHERE name IN (
    'audits:read',
    'questions:read',
    'submissions:read', 'submissions:create', 'submissions:update',
    'evidence:read', 'evidence:upload',
    'comments:read', 'comments:create'
);

-- Viewer: Read-only access
INSERT INTO client_role_permissions (role_id, permission_id)
SELECT '44444444-4444-4444-4444-444444444444', id FROM client_permissions
WHERE name IN (
    'audits:read',
    'questions:read',
    'submissions:read',
    'evidence:read',
    'reports:read',
    'comments:read'
);
