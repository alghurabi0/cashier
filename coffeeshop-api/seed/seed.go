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

	fmt.Println("✅ Seeding complete!")
}

func seedCategories(db *sqlx.DB) error {
	categories := []struct {
		NameAr    string
		SortOrder int
	}{
		{"مشروبات ساخنة", 1},
		{"شاي", 2},
		{"لاتيه", 3},
		{"كابتشينو", 4},
		{"آيس درنك", 5},
		{"آيس لاتيه", 6},
		{"فرابتشينو", 7},
		{"شيكات", 8},
		{"موهيتو", 9},
		{"آيس كريم", 10},
		{"كريب", 11},
		{"وافل", 12},
		{"VIP", 13},
	}

	for _, c := range categories {
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
		Description    string
		Price          int64
	}{
		// مشروبات ساخنة
		{"مشروبات ساخنة", "اسبريسو سينكل", "اسبريسو", 1500},
		{"مشروبات ساخنة", "اسبريسو دبل", "اسبريسو", 2000},
		{"مشروبات ساخنة", "امريكانو", "اسبريسو ، ماء ساخن", 2000},
		{"مشروبات ساخنة", "كورتادو", "اسبريسو ، حليب ساخن", 2500},
		{"مشروبات ساخنة", "هوت جوكليت", "شوكولاتة ، حليب", 3000},
		{"مشروبات ساخنة", "اسبرسو ماكياتو", "اسبريسو ، زغوة حليب", 2500},
		{"مشروبات ساخنة", "موكا", "اسبريسو ، حليب ، شوكولاتة", 4000},
		{"مشروبات ساخنة", "اسبريسو جكللتيه", "اسبريسو ، شوكولاتة", 2000},
		{"مشروبات ساخنة", "قهوة فرنسية", "قهوة ، ماء ساخن", 3000},
		{"مشروبات ساخنة", "حليب زعفران", "حليب ، زعفران", 2500},

		// شاي
		{"شاي", "شاي", "شاي سياه", 1000},
		{"شاي", "شاي كرك", "شاي سياه ، حليب ، سكر ، هيل", 3000},
		{"شاي", "شاي ماسالا", "شاي سياه ، حليب ، بهارات ماسالا ، سكر", 3000},
		{"شاي", "شاي زعفران", "شاي سياه ، زعفران", 1500},
		{"شاي", "شاي أعشاب", "أعشاب طبيعية", 2000},

		// لاتيه
		{"لاتيه", "لاتيه", "اسبريسو ، حليب", 4000},
		{"لاتيه", "لاتيه بندق", "اسبريسو ، حليب ، سيروب بندق", 4500},
		{"لاتيه", "لاتيه فانيلا", "اسبريسو ، حليب ، سيروب فانيلا", 4500},
		{"لاتيه", "لاتيه كارامل", "اسبريسو ، حليب ، سيروب كارامل", 4500},
		{"لاتيه", "لاتيه جوكليت", "اسبريسو ، حليب ، شوكولاتة", 4500},
		{"لاتيه", "لاتيه كوكونات", "اسبريسو ، حليب ، سيروب كوكونات", 4500},
		{"لاتيه", "لاتيه فراولة", "اسبريسو ، حليب ، سيروب فراولة", 4500},
		{"لاتيه", "لاتيه فستق", "اسبريسو ، حليب ، سيروب فستق", 4500},
		{"لاتيه", "لاتيه مانكو", "اسبريسو ، حليب ، سيروب مانكو", 4500},
		{"لاتيه", "لاتيه لوتوس", "اسبريسو ، حليب ، لوتوس", 4500},
		{"لاتيه", "لاتيه نوتلا", "اسبريسو ، حليب ، نوتلا", 4500},

		// كابتشينو
		{"كابتشينو", "كابتشينو", "اسبريسو ، حليب ، فوم الحليب", 3000},
		{"كابتشينو", "كابتشينو بندق", "اسبريسو ، حليب ، فوم الحليب ، بندق", 3500},
		{"كابتشينو", "كابتشينو فانيلا", "اسبريسو ، حليب ، فوم الحليب ، فانيلا", 3500},
		{"كابتشينو", "كابتشينو كارامل", "اسبريسو ، حليب ، فوم الحليب ، كارامل", 3500},
		{"كابتشينو", "كابتشينو جوكليت", "اسبريسو ، حليب ، فوم الحليب ، شوكولاتة", 3500},
		{"كابتشينو", "كابتشينو كوكونات", "اسبريسو ، حليب ، فوم الحليب ، كوكونات", 3500},

		// آيس درنك
		{"آيس درنك", "ايس دبل", "اسبريسو دبل ، ماء ، ثلج", 2500},
		{"آيس درنك", "كلد برو", "قهوة كلد برو ، ماء ، ثلج", 2500},
		{"آيس درنك", "ايس امريكانو", "اسبريسو ، ماء ، ثلج", 2500},
		{"آيس درنك", "ايس كارامل ماكياتو", "حليب ، اسبريسو ، صوص كراميل ، ثلج", 4500},
		{"آيس درنك", "ايس موكا", "حليب ، اسبريسو ، صوص شوكولاتة ، ثلج", 4500},
		{"آيس درنك", "ايس قهوة شوكليت", "حليب ، قهوة ، صوص شوكليت ، ثلج", 4000},
		{"آيس درنك", "ايس شوكولوس", "حليب ، شوكولاتة ، صوص شوكولاتة ، ثلج", 4500},
		{"آيس درنك", "ايس لوتوس موكا", "حليب ، اسبريسو ، صوص لوتوس ، ثلج", 4500},

		// آيس لاتيه
		{"آيس لاتيه", "ايس لاتيه", "اسبريسو ، حليب ، ثلج", 4500},
		{"آيس لاتيه", "إسبانيش لاتيه", "اسبريسو ، حليب مكثف ، ثلج", 5000},
		{"آيس لاتيه", "ايس لاتيه كارامل", "اسبريسو ، حليب ، صوص كارامل ، ثلج", 5000},
		{"آيس لاتيه", "ايس لاتية بندق", "اسبريسو ، حليب ، صوص بندق ، ثلج", 5000},
		{"آيس لاتيه", "ايس لاتية فانيلا", "اسبريسو ، حليب ، صوص فانيلا ، ثلج", 5000},
		{"آيس لاتيه", "ايس لاتية كوكونات", "اسبريسو ، حليب ، صوص كوكونات ، ثلج", 5000},

		// فرابتشينو
		{"فرابتشينو", "فرابتشينو كراميل", "كراميل ، حليب ، كريمة ، قهوة", 3500},
		{"فرابتشينو", "فرابتشينو فانيلا", "فانيلا ، حليب ، كريمة ، قهوة", 4000},
		{"فرابتشينو", "فرابتشينو نوتيلا", "نوتيلا ، حليب ، كريمة ، قهوة", 4000},
		{"فرابتشينو", "فرابتشينو فستق", "فستق ، حليب ، كريمة ، قهوة", 4000},
		{"فرابتشينو", "فرابتشينو كوكونات", "كوونات ، حليب ، كريمة ، قهوة", 3500},

		// شيكات
		{"شيكات", "ميلك شيك نوتلا", "حليب ، آيس كريم فانيليا ، نوتلا ، صوص نوتلا ، بندق مجروش", 5000},
		{"شيكات", "ميلك شيك لوتوس", "حليب ، آيس كريم فانيليا ، لوتوس ، صوص لوتوس ، بسكويت لوتوس", 5000},
		{"شيكات", "ميلك شيك فراوله", "حليب ، آيس كريم فانيليا ، فراولة ، صوص فراولة ، قطع فراولة", 4500},
		{"شيكات", "ميلك شيك موز نوتلا", "حليب ، آيس كريم فانيليا ، موز ، نوتلا ، صوص نوتلا ، شرائح موز", 5000},
		{"شيكات", "ميلك شيك موز فستق", "حليب ، آيس كريم فانيليا ، موز ، فستق ، صوص فستق ، فستق مجروش", 5000},

		// موهيتو
		{"موهيتو", "موهيتو", "نعناع طازج ، لايم ، شرابات سكر ، مياه فوازة", 3000},
		{"موهيتو", "موهيتو فراوله", "نعناع طازج ، فراولة ، لايم ، شرابات سكر ، مياه فوازة", 3000},
		{"موهيتو", "موهيتو خاص إن جي", "نعناع طازج ، توت أزرق ، زنجبيل ، شرابات سكر ، مياه فوازة", 3500},
		{"موهيتو", "بلو موهيتو", "نعناع طازج ، شرابات بلو كوراكاو ، شرابات سكر ، مياه فوازة", 3500},
		{"موهيتو", "مانكو موهيتو", "نعناع طازج ، مانجو ، لايم ، شرابات سكر ، مياه فوازة", 3500},
		{"موهيتو", "كيوي موهيتو", "نعناع طازج ، كيوي ، لايم ، شرابات سكر ، مياه فوازة", 3500},

		// آيس كريم
		{"آيس كريم", "أفوكاتو", "بستني فانيل ، إسبريسو", 3000},
		{"آيس كريم", "ثلاث سكويات", "اختياري من نكهات متنوعة", 2000},
		{"آيس كريم", "اربع سكويات", "اختياري من نكهات متنوعة", 2500},

		// كريب
		{"كريب", "كريب نوتلا", "نوتلا ، صوص شوكولاتة ، فستق ، كريمة", 4000},
		{"كريب", "كريب لوتوس", "صوص لوتوس ، بسكويت لوتوس ، كريمة", 4000},
		{"كريب", "كريب دبي شوكلييت", "شوكليت دبي ، صوص شوكولاتة ، فستق ، كنافة ، كريمة", 4500},

		// وافل
		{"وافل", "وافل نوتلا", "نوتلا", 4000},
		{"وافل", "وافل لوتوس", "لوتوس", 4000},
		{"وافل", "وافل ايس كريم", "آيس كريم", 4500},

		// VIP
		{"VIP", "صحن فواكه", "فواكه طازجة متنوعة", 7000},
		{"VIP", "ايس إن جي", "آيس كريم لوتوس خاص", 4500},
		{"VIP", "معجون إن جي", "معجون خاص بالكافيه", 7000},
		{"VIP", "كريب إن جي", "كريب خاص بالكافيه", 5500},
		{"VIP", "ماشا إن جي", "ماتشا خاص بالكافيه", 5000},
	}

	for _, item := range items {
		catID, ok := catMap[item.CategoryNameAr]
		if !ok {
			fmt.Printf("  ⚠ Category '%s' not found, skipping item '%s'\n", item.CategoryNameAr, item.NameAr)
			continue
		}

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
			`INSERT INTO menu_items (category_id, name_ar, description_ar, price) VALUES ($1, $2, $3, $4)`,
			catID, item.NameAr, item.Description, item.Price,
		)
		if err != nil {
			_, err = db.Exec(
				`INSERT INTO menu_items (category_id, name_ar, price) VALUES ($1, $2, $3)`,
				catID, item.NameAr, item.Price,
			)
			if err != nil {
				return fmt.Errorf("insert menu item '%s': %w", item.NameAr, err)
			}
		}
		fmt.Printf("  ✅ Created menu item: %s\n", item.NameAr)
	}

	return nil
}

func seedInventoryItems(db *sqlx.DB) error {
	items := []struct {
		NameAr     string
		BaseUnitAr string
		StockQty   int
		UnitCost   int64
	}{
		{"حبوب قهوة", "غرام", 5000, 3},
		{"حليب", "مل", 20000, 1},
		{"سكر", "غرام", 10000, 1},
		{"أكواب ورقية", "قطعة", 500, 50},
		{"شوكولاتة", "غرام", 3000, 5},
		{"فراولة مجمدة", "غرام", 2000, 4},
		{"لوتوس", "غرام", 1000, 10},
		{"نوتلا", "غرام", 1000, 8},
		{"كريمة", "مل", 3000, 3},
		{"فستق", "غرام", 1000, 15},
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
		fmt.Printf("  ✅ Created inventory item: %s\n", item.NameAr)
	}

	return nil
}
