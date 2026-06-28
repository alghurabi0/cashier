ALTER TABLE users DROP CONSTRAINT IF EXISTS users_tenant_username_unique;
ALTER TABLE tables DROP CONSTRAINT IF EXISTS tables_tenant_number_unique;

ALTER TABLE users DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE categories DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE menu_items DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE recipe_ingredients DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE orders DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE order_items DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE stock_adjustments DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE tables DROP COLUMN IF EXISTS tenant_id;

-- Restore original unique constraints
ALTER TABLE users ADD CONSTRAINT users_username_key UNIQUE (username);
ALTER TABLE tables ADD CONSTRAINT tables_number_key UNIQUE (number);
