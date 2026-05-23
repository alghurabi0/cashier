package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// ConnectSQLite opens (or creates) a SQLite database at the given path.
func ConnectSQLite(dbPath string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
	}

	// SQLite-specific pragmas for performance
	db.MustExec("PRAGMA foreign_keys = ON")

	return db, nil
}
