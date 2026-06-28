-- Rollback: remove updated_at columns and triggers

DROP TRIGGER IF EXISTS trg_recipe_ingredients_updated_at ON recipe_ingredients;
DROP TRIGGER IF EXISTS trg_inventory_items_updated_at ON inventory_items;
DROP TRIGGER IF EXISTS trg_menu_items_updated_at ON menu_items;
DROP TRIGGER IF EXISTS trg_categories_updated_at ON categories;

ALTER TABLE recipe_ingredients DROP COLUMN IF EXISTS updated_at;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS updated_at;
ALTER TABLE menu_items DROP COLUMN IF EXISTS updated_at;
ALTER TABLE categories DROP COLUMN IF EXISTS updated_at;

DROP FUNCTION IF EXISTS update_updated_at();
