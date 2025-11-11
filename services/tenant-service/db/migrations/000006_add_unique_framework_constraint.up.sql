-- Add unique constraint to prevent duplicate framework assignments to a client in an audit cycle
ALTER TABLE audit_cycle_frameworks
ADD CONSTRAINT unique_framework_per_client_cycle UNIQUE (audit_cycle_client_id, framework_id);
