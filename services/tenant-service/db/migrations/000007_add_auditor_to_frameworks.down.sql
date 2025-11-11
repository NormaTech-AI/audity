-- Remove auditor_id column from audit_cycle_frameworks table
DROP INDEX IF EXISTS idx_audit_cycle_frameworks_auditor_id;
ALTER TABLE audit_cycle_frameworks DROP COLUMN IF EXISTS auditor_id;
