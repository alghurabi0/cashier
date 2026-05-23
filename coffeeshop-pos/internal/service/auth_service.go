package service

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// LocalUser represents a user stored in local SQLite for PIN auth.
type LocalUser struct {
	ID     string `db:"id"       json:"id"`
	NameAr string `db:"name_ar"  json:"name_ar"`
	Role   string `db:"role"     json:"role"`
}

// localUserRow includes the pin hash for internal lookups.
type localUserRow struct {
	LocalUser
	PinHash string `db:"pin_hash"`
}

// AuthService handles local PIN-based authentication.
// It is Wails-bound so the Vue frontend can call its methods.
type AuthService struct {
	db          *sqlx.DB
	currentUser *LocalUser
	mu          sync.RWMutex
}

// NewAuthService creates a new AuthService.
func NewAuthService(db *sqlx.DB) *AuthService {
	return &AuthService{db: db}
}

// SeedDefaultAdmin creates a default admin user if no users exist.
// Called once at startup.
func (s *AuthService) SeedDefaultAdmin() {
	var count int
	err := s.db.Get(&count, `SELECT COUNT(*) FROM local_users`)
	if err != nil || count > 0 {
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("0000"), bcrypt.DefaultCost)
	if err != nil {
		slog.Warn("auth: failed to hash default PIN", "error", err)
		return
	}

	id := uuid.New().String()
	_, err = s.db.Exec(
		`INSERT INTO local_users (id, name_ar, pin_hash, role) VALUES (?, ?, ?, 'admin')`,
		id, "المدير", string(hash),
	)
	if err != nil {
		slog.Warn("auth: failed to seed default admin", "error", err)
		return
	}
	slog.Info("auth: default admin user seeded (PIN: 0000)")
}

// Login verifies the PIN and sets the current user session.
func (s *AuthService) Login(pin string) (*LocalUser, error) {
	if pin == "" {
		return nil, fmt.Errorf("PIN is required")
	}

	var users []localUserRow
	err := s.db.Select(&users, `SELECT id, name_ar, pin_hash, role FROM local_users`)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}

	for _, u := range users {
		if bcrypt.CompareHashAndPassword([]byte(u.PinHash), []byte(pin)) == nil {
			s.mu.Lock()
			s.currentUser = &LocalUser{ID: u.ID, NameAr: u.NameAr, Role: u.Role}
			s.mu.Unlock()
			slog.Info("auth: user logged in", "name", u.NameAr, "role", u.Role)
			return s.currentUser, nil
		}
	}

	return nil, fmt.Errorf("رمز PIN غير صحيح")
}

// GetCurrentUser returns the currently logged-in user, or nil.
func (s *AuthService) GetCurrentUser() *LocalUser {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentUser
}

// Logout clears the current user session.
func (s *AuthService) Logout() {
	s.mu.Lock()
	s.currentUser = nil
	s.mu.Unlock()
	slog.Info("auth: user logged out")
}

// HasUsers returns true if at least one user exists in the local_users table.
func (s *AuthService) HasUsers() bool {
	var count int
	s.db.Get(&count, `SELECT COUNT(*) FROM local_users`)
	return count > 0
}

// ── Admin-Only User Management ──

// CreateUser creates a new local user. Admin-only.
func (s *AuthService) CreateUser(nameAr, pin, role string) (*LocalUser, error) {
	if err := s.requireAdmin(); err != nil {
		return nil, err
	}
	if nameAr == "" || pin == "" {
		return nil, fmt.Errorf("name and PIN are required")
	}
	if role != "admin" && role != "cashier" {
		return nil, fmt.Errorf("role must be 'admin' or 'cashier'")
	}
	if len(pin) < 4 {
		return nil, fmt.Errorf("PIN must be at least 4 digits")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash PIN: %w", err)
	}

	id := uuid.New().String()
	_, err = s.db.Exec(
		`INSERT INTO local_users (id, name_ar, pin_hash, role) VALUES (?, ?, ?, ?)`,
		id, nameAr, string(hash), role,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &LocalUser{ID: id, NameAr: nameAr, Role: role}, nil
}

// ListUsers returns all local users (without PIN hash).
func (s *AuthService) ListUsers() ([]LocalUser, error) {
	var users []LocalUser
	err := s.db.Select(&users, `SELECT id, name_ar, role FROM local_users ORDER BY name_ar ASC`)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

// DeleteUser removes a local user. Admin-only. Cannot delete self.
func (s *AuthService) DeleteUser(id string) error {
	if err := s.requireAdmin(); err != nil {
		return err
	}

	s.mu.RLock()
	currentID := ""
	if s.currentUser != nil {
		currentID = s.currentUser.ID
	}
	s.mu.RUnlock()

	if id == currentID {
		return fmt.Errorf("cannot delete yourself")
	}

	result, err := s.db.Exec(`DELETE FROM local_users WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// ChangePin updates a user's PIN. Admin-only.
func (s *AuthService) ChangePin(userID, newPin string) error {
	if err := s.requireAdmin(); err != nil {
		return err
	}
	if len(newPin) < 4 {
		return fmt.Errorf("PIN must be at least 4 digits")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPin), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash PIN: %w", err)
	}

	result, err := s.db.Exec(`UPDATE local_users SET pin_hash = ? WHERE id = ?`, string(hash), userID)
	if err != nil {
		return fmt.Errorf("failed to update PIN: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// requireAdmin checks that the current user has the admin role.
func (s *AuthService) requireAdmin() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.currentUser == nil {
		return fmt.Errorf("not logged in")
	}
	if s.currentUser.Role != "admin" {
		return fmt.Errorf("admin access required")
	}
	return nil
}
