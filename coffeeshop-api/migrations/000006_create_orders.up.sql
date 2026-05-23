CREATE TABLE orders (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number   TEXT UNIQUE,
    source         TEXT NOT NULL,
    table_number   TEXT,
    status         TEXT NOT NULL DEFAULT 'pending',
    total          BIGINT NOT NULL,
    payment_method TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE order_items (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id         UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    menu_item_id     UUID REFERENCES menu_items(id),
    quantity         INT NOT NULL,
    unit_price       BIGINT NOT NULL,
    line_total       BIGINT NOT NULL,
    name_ar_snapshot TEXT NOT NULL
);

CREATE INDEX idx_order_items_order_id ON order_items(order_id);
