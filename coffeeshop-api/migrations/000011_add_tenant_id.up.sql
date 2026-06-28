-- Add tenant_id to all tenant-scoped tables.
-- No legacy data to migrate (dev-only), so NOT NULL from the start.

ALTER TABLE users ADD COLUMN tenant_id UUID NOT NULL REFERENCES tenants(id);
ALTER TABLE categories ADD COLUMN tenant_id UUID NOT NULL REFERENCES tenants(id);
ALTER TABLE menu_items ADD COLUMN tenant_id UUID NOT NULL REFERENCES tenants(id);
ALTER TABLE inventory_items ADD COLUMN tenant_id UUID NOT NULL REFERENCES tenants(id);
ALTER TABLE recipe_ingredients ADD COLUMN tenant_id UUID NOT NULL REFERENCES tenants(id);
ALTER TABLE orders ADD COLUMN tenant_id UUID NOT NULL REFERENCES tenants(id);
ALTER TABLE order_items ADD COLUMN tenant_id UUID NOT NULL REFERENCES tenants(id);
ALTER TABLE stock_adjustments ADD COLUMN tenant_id UUID NOT NULL REFERENCES tenants(id);
ALTER TABLE tables ADD COLUMN tenant_id UUID NOT NULL REFERENCES tenants(id);

-- Indexes for tenant scoping
CREATE INDEX idx_users_tenant ON users(tenant_id);
CREATE INDEX idx_categories_tenant ON categories(tenant_id);
CREATE INDEX idx_menu_items_tenant ON menu_items(tenant_id);
CREATE INDEX idx_inventory_items_tenant ON inventory_items(tenant_id);
CREATE INDEX idx_recipe_ingredients_tenant ON recipe_ingredients(tenant_id);
CREATE INDEX idx_orders_tenant ON orders(tenant_id);
CREATE INDEX idx_order_items_tenant ON order_items(tenant_id);
CREATE INDEX idx_stock_adjustments_tenant ON stock_adjustments(tenant_id);
CREATE INDEX idx_tables_tenant ON tables(tenant_id);

-- Username unique per-tenant (not globally)
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_key;
ALTER TABLE users ADD CONSTRAINT users_tenant_username_unique UNIQUE (tenant_id, username);

-- Table number unique per-tenant
ALTER TABLE tables DROP CONSTRAINT IF EXISTS tables_number_key;
ALTER TABLE tables ADD CONSTRAINT tables_tenant_number_unique UNIQUE (tenant_id, number);
