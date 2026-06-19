package service

import (
	"fmt"
	"log/slog"

	posSync "coffeeshop-pos/internal/sync"

	"github.com/jmoiron/sqlx"
)

// APIConnection holds the stored API connection details.
type APIConnection struct {
	APIURL   string `json:"api_url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// ConfigStoreService is a Wails-bound service for persistent app configuration.
// It stores key-value pairs in the local SQLite `app_config` table.
type ConfigStoreService struct {
	db        *sqlx.DB
	apiClient *posSync.APIClient
}

// NewConfigStoreService creates a new ConfigStoreService.
func NewConfigStoreService(db *sqlx.DB, apiClient *posSync.APIClient) *ConfigStoreService {
	return &ConfigStoreService{db: db, apiClient: apiClient}
}

// Get reads a config value by key.
func (s *ConfigStoreService) Get(key string) string {
	var val string
	err := s.db.Get(&val, `SELECT value FROM app_config WHERE key = ?`, key)
	if err != nil {
		return ""
	}
	return val
}

// Set writes a config value.
func (s *ConfigStoreService) Set(key, value string) error {
	_, err := s.db.Exec(
		`INSERT INTO app_config (key, value) VALUES (?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, value,
	)
	return err
}

// IsSetup returns true if the API connection has been configured.
func (s *ConfigStoreService) IsSetup() bool {
	return s.Get("api_url") != "" && s.Get("api_username") != ""
}

// GetAPIConnection returns the stored API connection details.
func (s *ConfigStoreService) GetAPIConnection() APIConnection {
	return APIConnection{
		APIURL:   s.Get("api_url"),
		Username: s.Get("api_username"),
		Password: s.Get("api_password"),
	}
}

// SetupAPIConnection validates the API connection, stores credentials, and logs in.
// This is called from the setup wizard and the settings panel.
func (s *ConfigStoreService) SetupAPIConnection(apiURL, username, password string) error {
	if apiURL == "" || username == "" || password == "" {
		return fmt.Errorf("جميع الحقول مطلوبة")
	}

	// Update the API client's base URL and attempt login
	s.apiClient.SetBaseURL(apiURL)
	if err := s.apiClient.Login(username, password); err != nil {
		return fmt.Errorf("فشل الاتصال: %w", err)
	}

	// Connection successful — persist credentials
	if err := s.Set("api_url", apiURL); err != nil {
		return fmt.Errorf("failed to save api_url: %w", err)
	}
	if err := s.Set("api_username", username); err != nil {
		return fmt.Errorf("failed to save username: %w", err)
	}
	if err := s.Set("api_password", password); err != nil {
		return fmt.Errorf("failed to save password: %w", err)
	}

	slog.Info("API connection configured", "url", apiURL, "user", username)
	return nil
}

// TryAutoLogin attempts to login using stored credentials.
// Returns true if login succeeded, false otherwise.
func (s *ConfigStoreService) TryAutoLogin() bool {
	conn := s.GetAPIConnection()
	if conn.APIURL == "" || conn.Username == "" {
		return false
	}

	s.apiClient.SetBaseURL(conn.APIURL)
	if err := s.apiClient.Login(conn.Username, conn.Password); err != nil {
		slog.Warn("auto-login failed", "url", conn.APIURL, "error", err)
		return false
	}

	slog.Info("auto-login successful", "url", conn.APIURL)
	return true
}

// IsKitchenModeEnabled returns whether the kitchen preparation step is active.
// Default is false (orders go directly to 'completed').
func (s *ConfigStoreService) IsKitchenModeEnabled() bool {
	return s.Get("kitchen_mode_enabled") == "true"
}

// SetKitchenModeEnabled updates the kitchen mode setting.
func (s *ConfigStoreService) SetKitchenModeEnabled(enabled bool) error {
	val := "false"
	if enabled {
		val = "true"
	}
	return s.Set("kitchen_mode_enabled", val)
}
