-- Remove unique constraint for framework assignments
ALTER TABLE audit_cycle_frameworks
DROP CONSTRAINT IF EXISTS unique_framework_per_client_cycle;
