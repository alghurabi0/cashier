CREATE TABLE app_release (
    id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    version TEXT NOT NULL DEFAULT '1.0.0',
    release_notes TEXT NOT NULL DEFAULT '',
    download_url_win TEXT NOT NULL DEFAULT '',
    download_url_mac TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO app_release (version) VALUES ('1.0.0');
