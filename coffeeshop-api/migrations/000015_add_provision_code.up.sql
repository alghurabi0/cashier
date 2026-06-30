ALTER TABLE tenants ADD COLUMN provision_code TEXT UNIQUE;
ALTER TABLE tenants ADD COLUMN provision_code_expires_at TIMESTAMPTZ;
