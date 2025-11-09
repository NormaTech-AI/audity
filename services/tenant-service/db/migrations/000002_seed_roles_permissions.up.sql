-- Insert default roles
INSERT INTO roles (id, name, description) VALUES
    ('11111111-1111-1111-1111-111111111111', 'nishaj_admin', 'System administrator with full access'),
    ('22222222-2222-2222-2222-222222222222', 'auditor', 'Primary reviewer who validates evidence'),
    ('33333333-3333-3333-3333-333333333333', 'team_member', 'Support staff for auditors'),
    ('44444444-4444-4444-4444-444444444444', 'poc_internal', 'Internal point of contact for client relationship'),
    ('55555555-5555-5555-5555-555555555555', 'poc_client', 'Client point of contact'),
    ('66666666-6666-6666-6666-666666666666', 'stakeholder', 'Client employee assigned specific questions');

-- Insert permissions for clients resource
INSERT INTO permissions (name, resource, action, description) VALUES
    ('clients:create', 'clients', 'create', 'Create new clients'),
    ('clients:read', 'clients', 'read', 'View client information'),
    ('clients:update', 'clients', 'update', 'Update client information'),
    ('clients:delete', 'clients', 'delete', 'Delete clients'),
    ('clients:list', 'clients', 'list', 'List all clients');

-- Insert permissions for frameworks resource
INSERT INTO permissions (name, resource, action, description) VALUES
    ('frameworks:create', 'frameworks', 'create', 'Create compliance frameworks'),
    ('frameworks:read', 'frameworks', 'read', 'View compliance frameworks'),
    ('frameworks:update', 'frameworks', 'update', 'Update compliance frameworks'),
    ('frameworks:delete', 'frameworks', 'delete', 'Delete compliance frameworks'),
    ('frameworks:list', 'frameworks', 'list', 'List all frameworks');

-- Insert permissions for audits resource
INSERT INTO permissions (name, resource, action, description) VALUES
    ('audits:create', 'audits', 'create', 'Create audit assignments'),
    ('audits:read', 'audits', 'read', 'View audit details'),
    ('audits:update', 'audits', 'update', 'Update audit status'),
    ('audits:delete', 'audits', 'delete', 'Delete audits'),
    ('audits:list', 'audits', 'list', 'List audits'),
    ('audits:review', 'audits', 'review', 'Review and approve/reject submissions');

-- Insert permissions for questions resource
INSERT INTO permissions (name, resource, action, description) VALUES
    ('questions:read', 'questions', 'read', 'View questions'),
    ('questions:answer', 'questions', 'answer', 'Answer questions'),
    ('questions:delegate', 'questions', 'delegate', 'Delegate questions to others');

-- Insert permissions for evidence resource
INSERT INTO permissions (name, resource, action, description) VALUES
    ('evidence:upload', 'evidence', 'upload', 'Upload evidence files'),
    ('evidence:read', 'evidence', 'read', 'View evidence files'),
    ('evidence:delete', 'evidence', 'delete', 'Delete evidence files');

-- Insert permissions for reports resource
INSERT INTO permissions (name, resource, action, description) VALUES
    ('reports:generate', 'reports', 'generate', 'Generate audit reports'),
    ('reports:read', 'reports', 'read', 'View audit reports'),
    ('reports:sign', 'reports', 'sign', 'Upload signed reports'),
    ('reports:download', 'reports', 'download', 'Download reports');

-- Insert permissions for users resource
INSERT INTO permissions (name, resource, action, description) VALUES
    ('users:create', 'users', 'create', 'Create users'),
    ('users:read', 'users', 'read', 'View user information'),
    ('users:update', 'users', 'update', 'Update user information'),
    ('users:delete', 'users', 'delete', 'Delete users'),
    ('users:list', 'users', 'list', 'List all users');

-- Assign permissions to nishaj_admin role (full access)
INSERT INTO role_permissions (role_id, permission_id)
SELECT '11111111-1111-1111-1111-111111111111', id FROM permissions;

-- Assign permissions to auditor role
INSERT INTO role_permissions (role_id, permission_id)
SELECT '22222222-2222-2222-2222-222222222222', id FROM permissions
WHERE name IN (
    'clients:read', 'clients:list',
    'frameworks:read', 'frameworks:list',
    'audits:read', 'audits:list', 'audits:review',
    'questions:read',
    'evidence:read',
    'reports:generate', 'reports:read', 'reports:sign', 'reports:download'
);

-- Assign permissions to team_member role
INSERT INTO role_permissions (role_id, permission_id)
SELECT '33333333-3333-3333-3333-333333333333', id FROM permissions
WHERE name IN (
    'clients:read',
    'frameworks:read',
    'audits:read',
    'questions:read',
    'evidence:read',
    'reports:read'
);

-- Assign permissions to poc_internal role
INSERT INTO role_permissions (role_id, permission_id)
SELECT '44444444-4444-4444-4444-444444444444', id FROM permissions
WHERE name IN (
    'clients:read', 'clients:list',
    'frameworks:read', 'frameworks:list',
    'audits:read', 'audits:list',
    'questions:read',
    'evidence:read',
    'reports:read', 'reports:download'
);

-- Assign permissions to poc_client role
INSERT INTO role_permissions (role_id, permission_id)
SELECT '55555555-5555-5555-5555-555555555555', id FROM permissions
WHERE name IN (
    'audits:read',
    'questions:read', 'questions:answer', 'questions:delegate',
    'evidence:upload', 'evidence:read', 'evidence:delete',
    'reports:read', 'reports:download'
);

-- Assign permissions to stakeholder role
INSERT INTO role_permissions (role_id, permission_id)
SELECT '66666666-6666-6666-6666-666666666666', id FROM permissions
WHERE name IN (
    'questions:read', 'questions:answer',
    'evidence:upload', 'evidence:read'
);
