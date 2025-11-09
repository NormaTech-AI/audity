-- Add email_domain column to clients table
-- This enforces that only users with matching email domains can access the client
ALTER TABLE clients ADD COLUMN email_domain VARCHAR(255);

-- Create index for faster lookups
CREATE INDEX idx_clients_email_domain ON clients(email_domain);

-- Add comment
COMMENT ON COLUMN clients.email_domain IS 'Allowed email domain for client users (e.g., bagaria.com). Only users with emails ending in @<email_domain> can access this client.';
