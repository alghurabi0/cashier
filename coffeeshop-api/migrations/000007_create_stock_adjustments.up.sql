CREATE TABLE stock_adjustments (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inventory_item_id UUID NOT NULL REFERENCES inventory_items(id),
    delta             INT NOT NULL,
    reason_ar         TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_stock_adjustments_inventory_item ON stock_adjustments(inventory_item_id);
