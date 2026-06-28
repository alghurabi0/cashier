-- Seed the platform tenant for super-admin accounts.
-- This tenant is used exclusively for platform-level administration.
INSERT INTO tenants (id, name, slug, is_active, settings)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'المنصة',
    'platform',
    true,
    '{"kitchen_mode_enabled": false, "conflict_resolution_mode": "last-write-wins"}'
)
ON CONFLICT (slug) DO NOTHING;

-- Create the default super-admin user.
-- Password: "admin123" (bcrypt hash) — CHANGE IN PRODUCTION.
INSERT INTO users (tenant_id, username, password_hash, role)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'superadmin',
    '$2a$10$y/YMRat8CBVq3IFlyFhD6eIZNGIFtGYiSJS8x7j3nsvjxqZHnsgwi',
    'super_admin'
)
ON CONFLICT DO NOTHING;
