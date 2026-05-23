CREATE TABLE inventory_items (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name_ar             TEXT NOT NULL,
    base_unit_ar        TEXT NOT NULL,
    stock_qty           INT NOT NULL DEFAULT 0,
    low_stock_threshold INT NOT NULL DEFAULT 0,
    unit_cost           BIGINT NOT NULL DEFAULT 0,
    is_active           BOOLEAN NOT NULL DEFAULT true
);
