DROP TRIGGER IF EXISTS trg_order_items_updated_at ON order_items;
ALTER TABLE order_items DROP COLUMN IF EXISTS updated_at;

DROP TRIGGER IF EXISTS trg_orders_updated_at ON orders;
DROP INDEX IF EXISTS idx_orders_updated_at;
ALTER TABLE orders DROP COLUMN IF EXISTS updated_at;
