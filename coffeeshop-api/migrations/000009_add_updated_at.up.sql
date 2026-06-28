-- Add updated_at to sync-target tables with auto-update trigger

-- Shared trigger function
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- categories
ALTER TABLE categories ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
CREATE TRIGGER trg_categories_updated_at
    BEFORE UPDATE ON categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- menu_items
ALTER TABLE menu_items ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
CREATE TRIGGER trg_menu_items_updated_at
    BEFORE UPDATE ON menu_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- inventory_items
ALTER TABLE inventory_items ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
CREATE TRIGGER trg_inventory_items_updated_at
    BEFORE UPDATE ON inventory_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- recipe_ingredients
ALTER TABLE recipe_ingredients ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
CREATE TRIGGER trg_recipe_ingredients_updated_at
    BEFORE UPDATE ON recipe_ingredients
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
