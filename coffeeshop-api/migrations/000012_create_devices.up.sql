-- Devices table: each POS terminal registered under a tenant.
CREATE TABLE devices (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL REFERENCES tenants(id),
    device_name  TEXT NOT NULL,
    device_type  TEXT NOT NULL DEFAULT 'pos',
    is_active    BOOLEAN NOT NULL DEFAULT true,
    last_seen_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_devices_tenant ON devices(tenant_id);

-- Track which POS device created each order
ALTER TABLE orders ADD COLUMN device_id UUID REFERENCES devices(id);
CREATE INDEX idx_orders_device ON orders(device_id);
