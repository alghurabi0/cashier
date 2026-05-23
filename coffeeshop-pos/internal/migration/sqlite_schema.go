package migration

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// RunSQLiteMigrations creates all tables in the local SQLite database.
// This mirrors the PostgreSQL schema but adapted for SQLite.
func RunSQLiteMigrations(db *sqlx.DB) error {
	tables := []struct {
		Name string
		SQL  string
	}{
		{"categories", `
			CREATE TABLE IF NOT EXISTS categories (
				id         TEXT PRIMARY KEY,
				name_ar    TEXT NOT NULL,
				sort_order INTEGER NOT NULL DEFAULT 0,
				is_active  INTEGER NOT NULL DEFAULT 1
			)`},
		{"menu_items", `
			CREATE TABLE IF NOT EXISTS menu_items (
				id                TEXT PRIMARY KEY,
				category_id       TEXT NOT NULL REFERENCES categories(id),
				name_ar           TEXT NOT NULL,
				price             INTEGER NOT NULL,
				cost_calc_method  TEXT NOT NULL DEFAULT 'auto',
				manual_cost_price INTEGER NOT NULL DEFAULT 0,
				cached_auto_cost  INTEGER NOT NULL DEFAULT 0,
				image_path        TEXT DEFAULT '',
				is_active         INTEGER NOT NULL DEFAULT 1
			)`},
		{"inventory_items", `
			CREATE TABLE IF NOT EXISTS inventory_items (
				id                  TEXT PRIMARY KEY,
				name_ar             TEXT NOT NULL,
				base_unit_ar        TEXT NOT NULL,
				stock_qty           INTEGER NOT NULL DEFAULT 0,
				low_stock_threshold INTEGER NOT NULL DEFAULT 0,
				unit_cost           INTEGER NOT NULL DEFAULT 0,
				is_active           INTEGER NOT NULL DEFAULT 1
			)`},
		{"recipe_ingredients", `
			CREATE TABLE IF NOT EXISTS recipe_ingredients (
				id                TEXT PRIMARY KEY,
				menu_item_id      TEXT NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
				inventory_item_id TEXT NOT NULL REFERENCES inventory_items(id) ON DELETE CASCADE,
				quantity          INTEGER NOT NULL
			)`},
		{"orders", `
			CREATE TABLE IF NOT EXISTS orders (
				id             TEXT PRIMARY KEY,
				order_number   TEXT UNIQUE,
				source         TEXT NOT NULL,
				table_number   TEXT,
				status         TEXT NOT NULL DEFAULT 'pending',
				total          INTEGER NOT NULL,
				payment_method TEXT,
				created_at     TEXT NOT NULL DEFAULT (datetime('now')),
				synced         INTEGER NOT NULL DEFAULT 0
			)`},
		{"order_items", `
			CREATE TABLE IF NOT EXISTS order_items (
				id               TEXT PRIMARY KEY,
				order_id         TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
				menu_item_id     TEXT,
				quantity         INTEGER NOT NULL,
				unit_price       INTEGER NOT NULL,
				line_total       INTEGER NOT NULL,
				name_ar_snapshot TEXT NOT NULL
			)`},
		{"stock_adjustments", `
			CREATE TABLE IF NOT EXISTS stock_adjustments (
				id                TEXT PRIMARY KEY,
				inventory_item_id TEXT NOT NULL REFERENCES inventory_items(id),
				delta             INTEGER NOT NULL,
				reason_ar         TEXT,
				created_at        TEXT NOT NULL DEFAULT (datetime('now')),
				synced            INTEGER NOT NULL DEFAULT 0
			)`},
		{"sync_meta", `
			CREATE TABLE IF NOT EXISTS sync_meta (
				table_name     TEXT PRIMARY KEY,
				last_synced_at TEXT NOT NULL DEFAULT ''
			)`},
		{"local_users", `
			CREATE TABLE IF NOT EXISTS local_users (
				id       TEXT PRIMARY KEY,
				name_ar  TEXT NOT NULL,
				pin_hash TEXT NOT NULL,
				role     TEXT NOT NULL DEFAULT 'cashier'
			)`},
		{"app_config", `
			CREATE TABLE IF NOT EXISTS app_config (
				key   TEXT PRIMARY KEY,
				value TEXT NOT NULL DEFAULT ''
			)`},
	}

	for _, t := range tables {
		if _, err := db.Exec(t.SQL); err != nil {
			return fmt.Errorf("failed to create table %s: %w", t.Name, err)
		}
	}

	return nil
}
