-- Add updated_at to orders table for delta-sync support
ALTER TABLE orders ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
CREATE TRIGGER trg_orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Add updated_at to order_items table
ALTER TABLE order_items ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
CREATE TRIGGER trg_order_items_updated_at
    BEFORE UPDATE ON order_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Index for efficient delta sync queries
CREATE INDEX idx_orders_updated_at ON orders(updated_at);
