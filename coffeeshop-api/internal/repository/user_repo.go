package repository

import (
	"coffeeshop-api/internal/model"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// UserRepository handles database operations for users.
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByUsername retrieves a user by username.
func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Get(&user, `SELECT id, username, password_hash, role, created_at FROM users WHERE username = $1`, username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

// FindByID retrieves a user by ID.
func (r *UserRepository) FindByID(id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.Get(&user, `SELECT id, username, password_hash, role, created_at FROM users WHERE id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

// Create inserts a new user and returns the created user.
func (r *UserRepository) Create(username, passwordHash, role string) (*model.User, error) {
	var user model.User
	err := r.db.Get(&user,
		`INSERT INTO users (username, password_hash, role) VALUES ($1, $2, $3)
		 RETURNING id, username, password_hash, role, created_at`,
		username, passwordHash, role,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return &user, nil
}

// Count returns the total number of users.
func (r *UserRepository) Count() (int, error) {
	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM users`)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}
