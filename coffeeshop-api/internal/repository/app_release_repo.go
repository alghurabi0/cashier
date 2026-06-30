package repository

import (
	"coffeeshop-api/internal/model"
	"time"

	"github.com/jmoiron/sqlx"
)

type AppReleaseRepository struct {
	db *sqlx.DB
}

func NewAppReleaseRepository(db *sqlx.DB) *AppReleaseRepository {
	return &AppReleaseRepository{db: db}
}

func (r *AppReleaseRepository) Get() (*model.AppRelease, error) {
	var release model.AppRelease
	err := r.db.Get(&release, `SELECT id, version, release_notes, download_url_win, download_url_mac, updated_at FROM app_release WHERE id = 1`)
	if err != nil {
		return nil, err
	}
	return &release, nil
}

func (r *AppReleaseRepository) Update(req *model.UpdateAppReleaseRequest) (*model.AppRelease, error) {
	var release model.AppRelease
	err := r.db.Get(&release,
		`UPDATE app_release
		 SET version = $1, release_notes = $2, download_url_win = $3, download_url_mac = $4, updated_at = $5
		 WHERE id = 1
		 RETURNING id, version, release_notes, download_url_win, download_url_mac, updated_at`,
		req.Version, req.ReleaseNotes, req.DownloadURLWin, req.DownloadURLMac, time.Now(),
	)
	if err != nil {
		return nil, err
	}
	return &release, nil
}
