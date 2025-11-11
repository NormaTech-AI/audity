-- Drop triggers
DROP TRIGGER IF EXISTS update_audit_cycle_frameworks_updated_at ON audit_cycle_frameworks;
DROP TRIGGER IF EXISTS update_audit_cycles_updated_at ON audit_cycles;

-- Drop tables in reverse order
DROP TABLE IF EXISTS audit_cycle_frameworks;
DROP TABLE IF EXISTS audit_cycle_clients;
DROP TABLE IF EXISTS audit_cycles;
