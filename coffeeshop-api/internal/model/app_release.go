package model

import "time"

type AppRelease struct {
	ID             int       `json:"id" db:"id"`
	Version        string    `json:"version" db:"version"`
	ReleaseNotes   string    `json:"release_notes" db:"release_notes"`
	DownloadURLWin string    `json:"download_url_win" db:"download_url_win"`
	DownloadURLMac string    `json:"download_url_mac" db:"download_url_mac"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type UpdateAppReleaseRequest struct {
	Version        string `json:"version"`
	ReleaseNotes   string `json:"release_notes"`
	DownloadURLWin string `json:"download_url_win"`
	DownloadURLMac string `json:"download_url_mac"`
}
