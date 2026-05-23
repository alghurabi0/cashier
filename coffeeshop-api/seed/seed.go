package main

import (
	"fmt"
	"log"

	"coffeeshop-api/internal/config"
	"coffeeshop-api/internal/database"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	fmt.Println("🌱 Seeding database...")

	if err := seedCategories(db); err != nil {
		log.Fatalf("failed to seed categories: %v", err)
	}

	if err := seedMenuItems(db); err != nil {
		log.Fatalf("failed to seed menu items: %v", err)
	}

	if err := seedInventoryItems(db); err != nil {
		log.Fatalf("failed to seed inventory items: %v", err)
	}

	if err := seedRecipes(db); err != nil {
		log.Fatalf("failed to seed recipes: %v", err)
	}

	fmt.Println("✅ Seeding complete!")
}

func seedCategories(db *sqlx.DB) error {
	categories := []struct {
		NameAr    string
		SortOrder int
	}{
		{"مشروبات ساخنة", 1},
		{"مشروبات باردة", 2},
		{"حلويات", 3},
	}

	for _, c := range categories {
		// Check if exists
		var count int
		err := db.Get(&count, `SELECT COUNT(*) FROM categories WHERE name_ar = $1`, c.NameAr)
		if err != nil {
			return fmt.Errorf("check category: %w", err)
		}
		if count > 0 {
			fmt.Printf("  ⏭ Category '%s' already exists, skipping\n", c.NameAr)
			continue
		}

		_, err = db.Exec(
			`INSERT INTO categories (name_ar, sort_order) VALUES ($1, $2)`,
			c.NameAr, c.SortOrder,
		)
		if err != nil {
			return fmt.Errorf("insert category '%s': %w", c.NameAr, err)
		}
		fmt.Printf("  ✅ Created category: %s\n", c.NameAr)
	}

	return nil
}

func seedMenuItems(db *sqlx.DB) error {
	// Fetch category IDs by name
	type catRow struct {
		ID     string `db:"id"`
		NameAr string `db:"name_ar"`
	}
	var cats []catRow
	if err := db.Select(&cats, `SELECT id, name_ar FROM categories WHERE is_active = true`); err != nil {
		return fmt.Errorf("fetch categories: %w", err)
	}

	catMap := make(map[string]string)
	for _, c := range cats {
		catMap[c.NameAr] = c.ID
	}

	items := []struct {
		CategoryNameAr string
		NameAr         string
		Price          int64 // fils (IQD × 1000)
	}{
		// مشروبات ساخنة (Hot drinks)
		{"مشروبات ساخنة", "اسبريسو", 2000},
		{"مشروبات ساخنة", "لاتيه", 3500},
		{"مشروبات ساخنة", "كابتشينو", 3500},
		{"مشروبات ساخنة", "شاي", 1500},

		// مشروبات باردة (Cold drinks)
		{"مشروبات باردة", "آيس لاتيه", 4000},
		{"مشروبات باردة", "آيس أمريكانو", 3500},
		{"مشروبات باردة", "سموذي فراولة", 4500},

		// حلويات (Sweets)
		{"حلويات", "كيك الشوكولاتة", 3000},
		{"حلويات", "تشيز كيك", 3500},
	}

	for _, item := range items {
		catID, ok := catMap[item.CategoryNameAr]
		if !ok {
			fmt.Printf("  ⚠ Category '%s' not found, skipping item '%s'\n", item.CategoryNameAr, item.NameAr)
			continue
		}

		// Check if exists
		var count int
		err := db.Get(&count, `SELECT COUNT(*) FROM menu_items WHERE name_ar = $1 AND category_id = $2`, item.NameAr, catID)
		if err != nil {
			return fmt.Errorf("check menu item: %w", err)
		}
		if count > 0 {
			fmt.Printf("  ⏭ Menu item '%s' already exists, skipping\n", item.NameAr)
			continue
		}

		_, err = db.Exec(
			`INSERT INTO menu_items (category_id, name_ar, price) VALUES ($1, $2, $3)`,
			catID, item.NameAr, item.Price,
		)
		if err != nil {
			return fmt.Errorf("insert menu item '%s': %w", item.NameAr, err)
		}
		fmt.Printf("  ✅ Created menu item: %s (%d fils)\n", item.NameAr, item.Price)
	}

	return nil
}

func seedInventoryItems(db *sqlx.DB) error {
	items := []struct {
		NameAr     string
		BaseUnitAr string
		StockQty   int
		UnitCost   int64 // fils per 1 base unit
	}{
		{"حبوب قهوة", "غرام", 5000, 3},       // coffee beans — 3 fils/g
		{"حليب", "مل", 20000, 1},              // milk — 1 fil/ml
		{"سكر", "غرام", 10000, 1},             // sugar — 1 fil/g
		{"أكواب ورقية", "قطعة", 500, 50},      // paper cups — 50 fils/piece
		{"شوكولاتة", "غرام", 3000, 5},         // chocolate — 5 fils/g
		{"فراولة مجمدة", "غرام", 2000, 4},     // frozen strawberry — 4 fils/g
	}

	for _, item := range items {
		var count int
		err := db.Get(&count, `SELECT COUNT(*) FROM inventory_items WHERE name_ar = $1`, item.NameAr)
		if err != nil {
			return fmt.Errorf("check inventory item: %w", err)
		}
		if count > 0 {
			fmt.Printf("  ⏭ Inventory item '%s' already exists, skipping\n", item.NameAr)
			continue
		}

		_, err = db.Exec(
			`INSERT INTO inventory_items (name_ar, base_unit_ar, stock_qty, unit_cost) VALUES ($1, $2, $3, $4)`,
			item.NameAr, item.BaseUnitAr, item.StockQty, item.UnitCost,
		)
		if err != nil {
			return fmt.Errorf("insert inventory item '%s': %w", item.NameAr, err)
		}
		fmt.Printf("  ✅ Created inventory item: %s (%s)\n", item.NameAr, item.BaseUnitAr)
	}

	return nil
}

func seedRecipes(db *sqlx.DB) error {
	// Helper to get ID by name
	type idRow struct {
		ID string `db:"id"`
	}

	getMenuItemID := func(nameAr string) (string, error) {
		var row idRow
		err := db.Get(&row, `SELECT id FROM menu_items WHERE name_ar = $1`, nameAr)
		if err != nil {
			return "", fmt.Errorf("menu item '%s' not found: %w", nameAr, err)
		}
		return row.ID, nil
	}

	getInventoryItemID := func(nameAr string) (string, error) {
		var row idRow
		err := db.Get(&row, `SELECT id FROM inventory_items WHERE name_ar = $1`, nameAr)
		if err != nil {
			return "", fmt.Errorf("inventory item '%s' not found: %w", nameAr, err)
		}
		return row.ID, nil
	}

	// Recipe definitions: menu item name → list of (inventory item name, quantity)
	recipes := []struct {
		MenuItemNameAr string
		Ingredients    []struct {
			InventoryNameAr string
			Quantity        int
		}
	}{
		{"اسبريسو", []struct {
			InventoryNameAr string
			Quantity        int
		}{
			{"حبوب قهوة", 18},    // 18g coffee
			{"أكواب ورقية", 1},   // 1 cup
		}},
		{"لاتيه", []struct {
			InventoryNameAr string
			Quantity        int
		}{
			{"حبوب قهوة", 18},    // 18g coffee
			{"حليب", 250},        // 250ml milk
			{"أكواب ورقية", 1},   // 1 cup
		}},
		{"كابتشينو", []struct {
			InventoryNameAr string
			Quantity        int
		}{
			{"حبوب قهوة", 18},    // 18g coffee
			{"حليب", 150},        // 150ml milk
			{"أكواب ورقية", 1},   // 1 cup
		}},
		{"آيس لاتيه", []struct {
			InventoryNameAr string
			Quantity        int
		}{
			{"حبوب قهوة", 18},    // 18g coffee
			{"حليب", 300},        // 300ml milk
			{"أكواب ورقية", 1},   // 1 cup
		}},
		{"آيس أمريكانو", []struct {
			InventoryNameAr string
			Quantity        int
		}{
			{"حبوب قهوة", 18},    // 18g coffee
			{"أكواب ورقية", 1},   // 1 cup
		}},
		{"سموذي فراولة", []struct {
			InventoryNameAr string
			Quantity        int
		}{
			{"فراولة مجمدة", 150}, // 150g frozen strawberry
			{"حليب", 200},        // 200ml milk
			{"سكر", 20},          // 20g sugar
			{"أكواب ورقية", 1},   // 1 cup
		}},
		{"كيك الشوكولاتة", []struct {
			InventoryNameAr string
			Quantity        int
		}{
			{"شوكولاتة", 50},     // 50g chocolate
			{"سكر", 30},          // 30g sugar
		}},
	}

	for _, recipe := range recipes {
		menuItemID, err := getMenuItemID(recipe.MenuItemNameAr)
		if err != nil {
			fmt.Printf("  ⚠ %v, skipping recipe\n", err)
			continue
		}

		// Check if recipe already exists
		var count int
		err = db.Get(&count, `SELECT COUNT(*) FROM recipe_ingredients WHERE menu_item_id = $1`, menuItemID)
		if err != nil {
			return fmt.Errorf("check recipe: %w", err)
		}
		if count > 0 {
			fmt.Printf("  ⏭ Recipe for '%s' already exists, skipping\n", recipe.MenuItemNameAr)
			continue
		}

		for _, ing := range recipe.Ingredients {
			inventoryItemID, err := getInventoryItemID(ing.InventoryNameAr)
			if err != nil {
				fmt.Printf("  ⚠ %v, skipping ingredient\n", err)
				continue
			}

			_, err = db.Exec(
				`INSERT INTO recipe_ingredients (menu_item_id, inventory_item_id, quantity) VALUES ($1, $2, $3)`,
				menuItemID, inventoryItemID, ing.Quantity,
			)
			if err != nil {
				return fmt.Errorf("insert recipe ingredient: %w", err)
			}
		}

		// Recalculate auto-cost
		var cost *int64
		err = db.Get(&cost,
			`SELECT SUM(ri.quantity::BIGINT * ii.unit_cost)
			 FROM recipe_ingredients ri
			 JOIN inventory_items ii ON ii.id = ri.inventory_item_id
			 WHERE ri.menu_item_id = $1`, menuItemID)
		if err != nil {
			return fmt.Errorf("calculate auto cost: %w", err)
		}
		if cost != nil {
			_, err = db.Exec(`UPDATE menu_items SET cached_auto_cost = $1 WHERE id = $2`, *cost, menuItemID)
			if err != nil {
				return fmt.Errorf("update cached auto cost: %w", err)
			}
		}

		fmt.Printf("  ✅ Created recipe for: %s (%d ingredients)\n", recipe.MenuItemNameAr, len(recipe.Ingredients))
	}

	return nil
}
