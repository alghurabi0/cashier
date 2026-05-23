CREATE TABLE recipe_ingredients (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    menu_item_id      UUID NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    inventory_item_id UUID NOT NULL REFERENCES inventory_items(id) ON DELETE CASCADE,
    quantity          INT NOT NULL
);

CREATE INDEX idx_recipe_ingredients_menu_item ON recipe_ingredients(menu_item_id);
CREATE INDEX idx_recipe_ingredients_inventory_item ON recipe_ingredients(inventory_item_id);
