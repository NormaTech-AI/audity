-- Remove email_domain column from clients table
DROP INDEX IF EXISTS idx_clients_email_domain;
ALTER TABLE clients DROP COLUMN IF EXISTS email_domain;
