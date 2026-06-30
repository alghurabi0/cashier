package config

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

type Config struct {
	APIBaseURL    string
	DatabasePath  string
	ImageCacheDir string
	SyncInterval  int
}

func Load() *Config {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = defaultDBPath()
	}

	imgDir := filepath.Join(filepath.Dir(dbPath), "images")
	os.MkdirAll(imgDir, 0755)

	return &Config{
		APIBaseURL:    getEnv("API_BASE_URL", "http://localhost:8080"),
		DatabasePath:  dbPath,
		ImageCacheDir: imgDir,
		SyncInterval:  30,
	}
}

func defaultDBPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "coffeeshop.db"
	}
	appDir := filepath.Join(configDir, "Zawan", "CashierPOS")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "coffeeshop.db"
	}

	target := filepath.Join(appDir, "coffeeshop.db")

	if _, err := os.Stat(target); os.IsNotExist(err) {
		if src, err := os.Stat("coffeeshop.db"); err == nil && !src.IsDir() {
			migrateDB("coffeeshop.db", target)
		}
	}

	return target
}

func migrateDB(src, dst string) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		os.Remove(dst)
		return
	}

	walSrc := src + "-wal"
	if _, err := os.Stat(walSrc); err == nil {
		walIn, err := os.Open(walSrc)
		if err == nil {
			walOut, err := os.Create(dst + "-wal")
			if err == nil {
				io.Copy(walOut, walIn)
				walOut.Close()
			}
			walIn.Close()
		}
	}

	shmSrc := src + "-shm"
	if _, err := os.Stat(shmSrc); err == nil {
		shmIn, err := os.Open(shmSrc)
		if err == nil {
			shmOut, err := os.Create(dst + "-shm")
			if err == nil {
				io.Copy(shmOut, shmIn)
				shmOut.Close()
			}
			shmIn.Close()
		}
	}

	slog.Info("migrated database to user data directory", "from", src, "to", dst)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
