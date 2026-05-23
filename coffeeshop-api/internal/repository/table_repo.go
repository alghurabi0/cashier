package repository

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"coffeeshop-api/internal/model"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// TableRepository handles database operations for tables.
type TableRepository struct {
	db *sqlx.DB
}

// NewTableRepository creates a new TableRepository.
func NewTableRepository(db *sqlx.DB) *TableRepository {
	return &TableRepository{db: db}
}

// List returns all active tables.
func (r *TableRepository) List() ([]model.Table, error) {
	var tables []model.Table
	err := r.db.Select(&tables,
		`SELECT id, number, token, is_active, created_at
		 FROM tables WHERE is_active = true ORDER BY number ASC`)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}
	return tables, nil
}

// GetByToken returns a table matching the given token, or nil if not found.
func (r *TableRepository) GetByToken(token string) (*model.Table, error) {
	var t model.Table
	err := r.db.Get(&t,
		`SELECT id, number, token, is_active, created_at
		 FROM tables WHERE token = $1 AND is_active = true`, token)
	if err != nil {
		return nil, fmt.Errorf("table not found for token: %w", err)
	}
	return &t, nil
}

// Create inserts a new table with an auto-generated token.
func (r *TableRepository) Create(number string) (*model.Table, error) {
	id := uuid.New()
	token, err := generateToken(12)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	var t model.Table
	err = r.db.Get(&t,
		`INSERT INTO tables (id, number, token) VALUES ($1, $2, $3)
		 RETURNING id, number, token, is_active, created_at`,
		id, number, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}
	return &t, nil
}

// Delete soft-deletes a table by ID.
func (r *TableRepository) Delete(id uuid.UUID) error {
	result, err := r.db.Exec(
		`UPDATE tables SET is_active = false WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete table: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("table not found")
	}
	return nil
}

// generateToken creates a cryptographically random hex token.
func generateToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b)[:length], nil
}
