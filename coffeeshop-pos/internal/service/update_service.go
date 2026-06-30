package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"coffeeshop-pos/internal/version"
)

type UpdateInfo struct {
	Available    bool   `json:"available"`
	Version      string `json:"version"`
	ReleaseNotes string `json:"release_notes"`
	DownloadURL  string `json:"download_url"`
}

type UpdateService struct {
	configStore *ConfigStoreService
	progress    atomic.Int64
	httpClient  *http.Client
}

func NewUpdateService(configStore *ConfigStoreService) *UpdateService {
	return &UpdateService{
		configStore: configStore,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *UpdateService) CheckForUpdate() UpdateInfo {
	apiURL := s.configStore.Get("api_url")
	if apiURL == "" {
		return UpdateInfo{}
	}

	url := strings.TrimRight(apiURL, "/") + "/api/v1/app/latest-version"
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return UpdateInfo{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return UpdateInfo{}
	}

	var data struct {
		Version      string `json:"version"`
		ReleaseNotes string `json:"release_notes"`
		DownloadWin  string `json:"download_url_win"`
		DownloadMac  string `json:"download_url_mac"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return UpdateInfo{}
	}

	if !isNewer(data.Version, version.Version) {
		return UpdateInfo{Version: data.Version}
	}

	downloadURL := data.DownloadWin
	if runtime.GOOS == "darwin" {
		downloadURL = data.DownloadMac
	}

	return UpdateInfo{
		Available:    true,
		Version:      data.Version,
		ReleaseNotes: data.ReleaseNotes,
		DownloadURL:  downloadURL,
	}
}

func (s *UpdateService) GetDownloadProgress() float64 {
	return float64(s.progress.Load()) / 100.0
}

func (s *UpdateService) DownloadUpdate(downloadURL string) (string, error) {
	if downloadURL == "" {
		return "", fmt.Errorf("no download URL")
	}

	dlClient := &http.Client{Timeout: 10 * time.Minute}
	resp, err := dlClient.Get(downloadURL)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	updateDir := filepath.Join(os.TempDir(), "cashier-update")
	os.MkdirAll(updateDir, 0755)

	ext := ".exe"
	if runtime.GOOS == "darwin" {
		ext = ".zip"
	}
	outPath := filepath.Join(updateDir, "cashier-update"+ext)

	out, err := os.Create(outPath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	totalSize := resp.ContentLength
	var written int64
	buf := make([]byte, 32*1024)

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return "", fmt.Errorf("write file: %w", writeErr)
			}
			written += int64(n)
			if totalSize > 0 {
				pct := int64(float64(written) / float64(totalSize) * 100)
				s.progress.Store(pct)
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return "", fmt.Errorf("read body: %w", readErr)
		}
	}

	s.progress.Store(100)
	return outPath, nil
}

func (s *UpdateService) ApplyUpdate(installerPath string) error {
	if runtime.GOOS == "windows" {
		return s.applyWindows(installerPath)
	}
	return fmt.Errorf("auto-update not supported on %s — please download manually", runtime.GOOS)
}

func (s *UpdateService) applyWindows(installerPath string) error {
	script := fmt.Sprintf("@echo off\r\ntimeout /t 2 /nobreak >nul\r\nstart \"\" \"%s\" /S\r\n", installerPath)
	scriptPath := filepath.Join(os.TempDir(), "cashier_update.bat")
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		return fmt.Errorf("write update script: %w", err)
	}

	cmd := exec.Command("cmd", "/c", "start", "", scriptPath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("launch update script: %w", err)
	}

	os.Exit(0)
	return nil
}

func isNewer(remote, local string) bool {
	rParts := parseVersion(remote)
	lParts := parseVersion(local)
	for i := 0; i < 3; i++ {
		if rParts[i] > lParts[i] {
			return true
		}
		if rParts[i] < lParts[i] {
			return false
		}
	}
	return false
}

func parseVersion(v string) [3]int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.SplitN(v, ".", 3)
	var result [3]int
	for i := 0; i < len(parts) && i < 3; i++ {
		n, _ := strconv.Atoi(strings.TrimRight(parts[i], "-abcdefghijklmnopqrstuvwxyz"))
		result[i] = n
	}
	return result
}
