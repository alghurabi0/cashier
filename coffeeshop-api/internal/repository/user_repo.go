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

// FindByUsername retrieves a user by username within a tenant.
func (r *UserRepository) FindByUsername(tenantID uuid.UUID, username string) (*model.User, error) {
	var user model.User
	err := r.db.Get(&user,
		`SELECT id, tenant_id, username, password_hash, role, created_at
		 FROM users WHERE tenant_id = $1 AND username = $2`,
		tenantID, username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

// FindByID retrieves a user by ID.
func (r *UserRepository) FindByID(id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.Get(&user,
		`SELECT id, tenant_id, username, password_hash, role, created_at
		 FROM users WHERE id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

// Create inserts a new user under a tenant and returns the created user.
func (r *UserRepository) Create(tenantID uuid.UUID, username, passwordHash, role string) (*model.User, error) {
	var user model.User
	err := r.db.Get(&user,
		`INSERT INTO users (tenant_id, username, password_hash, role) VALUES ($1, $2, $3, $4)
		 RETURNING id, tenant_id, username, password_hash, role, created_at`,
		tenantID, username, passwordHash, role,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return &user, nil
}

// Count returns the total number of users in a tenant.
func (r *UserRepository) Count(tenantID uuid.UUID) (int, error) {
	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM users WHERE tenant_id = $1`, tenantID)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}

// ListByTenant returns all users for a tenant.
func (r *UserRepository) ListByTenant(tenantID uuid.UUID) ([]model.User, error) {
	var users []model.User
	err := r.db.Select(&users,
		`SELECT id, tenant_id, username, role, created_at
		 FROM users WHERE tenant_id = $1 ORDER BY username ASC`,
		tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	if users == nil {
		users = []model.User{}
	}
	return users, nil
}
