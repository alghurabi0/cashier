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
	Username string `json:"username"` // format: "user@tenant-slug"
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
// Username should be in the format "user@tenant-slug".
func (s *ConfigStoreService) SetupAPIConnection(apiURL, username, password string) error {
	if apiURL == "" || username == "" || password == "" {
		return fmt.Errorf("جميع الحقول مطلوبة")
	}

	// Update the API client's base URL and attempt login
	s.apiClient.SetBaseURL(apiURL)
	loginResp, err := s.apiClient.Login(username, password)
	if err != nil {
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

	// Store tenant info from login response
	if loginResp != nil {
		s.Set("tenant_id", loginResp.Tenant.ID)
		s.Set("tenant_name", loginResp.Tenant.Name)
		s.Set("tenant_slug", loginResp.Tenant.Slug)

		// Sync tenant settings to local config
		if loginResp.Tenant.Settings.KitchenModeEnabled {
			s.Set("kitchen_mode_enabled", "true")
		} else {
			s.Set("kitchen_mode_enabled", "false")
		}
	}

	// Set device ID on the client if we have one stored
	if deviceID := s.Get("device_id"); deviceID != "" {
		s.apiClient.SetDeviceID(deviceID)
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
	loginResp, err := s.apiClient.Login(conn.Username, conn.Password)
	if err != nil {
		slog.Warn("auto-login failed", "url", conn.APIURL, "error", err)
		return false
	}

	// Update tenant settings on each login
	if loginResp != nil {
		s.Set("tenant_id", loginResp.Tenant.ID)
		if loginResp.Tenant.Settings.KitchenModeEnabled {
			s.Set("kitchen_mode_enabled", "true")
		} else {
			s.Set("kitchen_mode_enabled", "false")
		}
	}

	// Set device ID on the client
	if deviceID := s.Get("device_id"); deviceID != "" {
		s.apiClient.SetDeviceID(deviceID)
	}

	slog.Info("auto-login successful", "url", conn.APIURL)
	return true
}

// GetDeviceID returns the stored device ID for this POS instance.
func (s *ConfigStoreService) GetDeviceID() string {
	return s.Get("device_id")
}

// SetDeviceID stores the device ID and updates the API client.
func (s *ConfigStoreService) SetDeviceID(deviceID string) error {
	if err := s.Set("device_id", deviceID); err != nil {
		return err
	}
	s.apiClient.SetDeviceID(deviceID)
	return nil
}

// GetTenantID returns the stored tenant ID.
func (s *ConfigStoreService) GetTenantID() string {
	return s.Get("tenant_id")
}

// GetTenantName returns the stored tenant display name.
func (s *ConfigStoreService) GetTenantName() string {
	return s.Get("tenant_name")
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
