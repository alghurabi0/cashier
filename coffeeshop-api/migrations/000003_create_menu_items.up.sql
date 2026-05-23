CREATE TABLE menu_items (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id       UUID NOT NULL REFERENCES categories(id),
    name_ar           TEXT NOT NULL,
    price             BIGINT NOT NULL,
    cost_calc_method  TEXT NOT NULL DEFAULT 'auto',
    manual_cost_price BIGINT NOT NULL DEFAULT 0,
    cached_auto_cost  BIGINT NOT NULL DEFAULT 0,
    image_path        TEXT DEFAULT '',
    is_active         BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX idx_menu_items_category_id ON menu_items(category_id);
