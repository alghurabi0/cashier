package model

// PublicMenuResponse is the response for GET /api/v1/public/menu?token=xxx.
// Contains everything a standalone menu site needs in a single round-trip.
type PublicMenuResponse struct {
	Tenant     PublicTenantInfo `json:"tenant"`
	Table      PublicTableInfo  `json:"table"`
	Categories []Category       `json:"categories"`
	MenuItems  []MenuItem       `json:"menu_items"`
}

// PublicTenantInfo is the minimal tenant info exposed to public menu sites.
type PublicTenantInfo struct {
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	IntroVideoURL string `json:"intro_video_url,omitempty"`
}

// PublicTableInfo is the minimal table info exposed to public menu sites.
type PublicTableInfo struct {
	Number string `json:"number"`
}
