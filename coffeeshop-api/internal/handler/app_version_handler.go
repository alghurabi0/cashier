package handler

import (
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/repository"
	"net/http"
)

type AppVersionHandler struct {
	releaseRepo *repository.AppReleaseRepository
}

func NewAppVersionHandler(releaseRepo *repository.AppReleaseRepository) *AppVersionHandler {
	return &AppVersionHandler{releaseRepo: releaseRepo}
}

func (h *AppVersionHandler) LatestVersion(w http.ResponseWriter, r *http.Request) {
	release, err := h.releaseRepo.Get()
	if err != nil {
		JSON(w, http.StatusOK, map[string]any{
			"version":          "1.0.0",
			"release_notes":    "",
			"download_url_win": "",
			"download_url_mac": "",
		})
		return
	}

	JSON(w, http.StatusOK, map[string]any{
		"version":          release.Version,
		"release_notes":    release.ReleaseNotes,
		"download_url_win": release.DownloadURLWin,
		"download_url_mac": release.DownloadURLMac,
	})
}

func (h *AppVersionHandler) GetRelease(w http.ResponseWriter, r *http.Request) {
	release, err := h.releaseRepo.Get()
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, release)
}

func (h *AppVersionHandler) UpdateRelease(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateAppReleaseRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	if req.Version == "" {
		Error(w, http.StatusBadRequest, "version is required")
		return
	}

	release, err := h.releaseRepo.Update(&req)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, release)
}
