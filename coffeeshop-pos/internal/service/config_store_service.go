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
		s.syncTenantSettings(loginResp)
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

	if loginResp != nil {
		s.syncTenantSettings(loginResp)
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

// ProvisionWithCode provisions this POS using a setup code.
// Calls the provision endpoint, stores all returned config, and logs in.
func (s *ConfigStoreService) ProvisionWithCode(apiURL, code string) error {
	if apiURL == "" || code == "" {
		return fmt.Errorf("رابط الخادم ورمز الإعداد مطلوبان")
	}

	s.apiClient.SetBaseURL(apiURL)
	result, err := s.apiClient.Provision(code)
	if err != nil {
		return fmt.Errorf("فشل التفعيل: %w", err)
	}

	// Store credentials
	s.Set("api_url", apiURL)
	s.Set("api_username", result.Username)
	s.Set("api_password", result.Password)

	// Store tenant info
	s.Set("tenant_id", result.Tenant.ID)
	s.Set("tenant_name", result.Tenant.Name)
	s.Set("tenant_slug", result.Tenant.Slug)

	// Store settings
	if result.Tenant.Settings.KitchenModeEnabled {
		s.Set("kitchen_mode_enabled", "true")
	} else {
		s.Set("kitchen_mode_enabled", "false")
	}
	if result.Tenant.Settings.MenuURL != "" {
		s.Set("menu_url", result.Tenant.Settings.MenuURL)
	}
	if result.Tenant.Settings.IntroVideoURL != "" {
		s.Set("intro_video_url", result.Tenant.Settings.IntroVideoURL)
	}

	if deviceID := s.Get("device_id"); deviceID != "" {
		s.apiClient.SetDeviceID(deviceID)
	}

	slog.Info("POS provisioned", "tenant", result.Tenant.Slug, "url", apiURL)
	return nil
}

// syncTenantSettings stores tenant info from a login response.
func (s *ConfigStoreService) syncTenantSettings(resp *posSync.LoginResponse) {
	s.Set("tenant_id", resp.Tenant.ID)
	s.Set("tenant_name", resp.Tenant.Name)
	s.Set("tenant_slug", resp.Tenant.Slug)
	if resp.Tenant.Settings.KitchenModeEnabled {
		s.Set("kitchen_mode_enabled", "true")
	} else {
		s.Set("kitchen_mode_enabled", "false")
	}
	if resp.Tenant.Settings.MenuURL != "" {
		s.Set("menu_url", resp.Tenant.Settings.MenuURL)
	}
	if resp.Tenant.Settings.IntroVideoURL != "" {
		s.Set("intro_video_url", resp.Tenant.Settings.IntroVideoURL)
	}
}
