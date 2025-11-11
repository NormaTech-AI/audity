-- Add auditor_id column to audit_cycle_frameworks table
ALTER TABLE audit_cycle_frameworks
ADD COLUMN auditor_id UUID REFERENCES users(id) ON DELETE SET NULL;

-- Add index for auditor lookups
CREATE INDEX idx_audit_cycle_frameworks_auditor_id ON audit_cycle_frameworks(auditor_id);

-- Add comment to explain the column
COMMENT ON COLUMN audit_cycle_frameworks.auditor_id IS 'The auditor assigned to review this framework for the client';
