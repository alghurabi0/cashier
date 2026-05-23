package main

import (
	"context"
	"embed"
	"log"
	"log/slog"
	"os"

	"coffeeshop-pos/internal/config"
	"coffeeshop-pos/internal/database"
	"coffeeshop-pos/internal/migration"
	"coffeeshop-pos/internal/service"
	posSync "coffeeshop-pos/internal/sync"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Load config
	cfg := config.Load()

	// Setup logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	// Initialize SQLite
	db, err := database.ConnectSQLite(cfg.DatabasePath)
	if err != nil {
		slog.Error("failed to connect to SQLite", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("connected to SQLite", "path", cfg.DatabasePath)

	// Run migrations
	if err := migration.RunSQLiteMigrations(db); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("SQLite migrations complete")

	// Create services
	dataService := service.NewDataService(db)
	orderService := service.NewOrderService(db)
	receiptService := service.NewReceiptService("المقهى")
	authService := service.NewAuthService(db)
	reportService := service.NewReportService(db)

	// Seed default admin if no users exist (first launch)
	authService.SeedDefaultAdmin()

	// Create sync client and worker
	apiClient := posSync.NewAPIClient(cfg.APIBaseURL)
	syncWorker := posSync.NewWorker(apiClient, db)

	// ConfigStore: persistent API connection (replaces env var approach)
	configStore := service.NewConfigStoreService(db, apiClient)

	// Try auto-login using stored credentials (from previous setup)
	configStore.TryAutoLogin()

	// Management service (requires apiClient + syncWorker)
	managementService := service.NewManagementService(apiClient, syncWorker)

	// Web order service (manages incoming web menu orders)
	webOrderService := service.NewWebOrderService(db, apiClient)

	// SSE client for real-time web order notifications
	sseClient := posSync.NewSSEClient(cfg.APIBaseURL, func(event posSync.SSEEvent) {
		webOrderService.HandleSSEEvent(event)
	})

	// Start sync worker and SSE client in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go syncWorker.Start(ctx, cfg.SyncInterval)
	go sseClient.Connect(ctx)

	// Create Wails application
	app := application.New(application.Options{
		Name:        "Coffeeshop POS",
		Description: "نقطة البيع - المقهى",
		Services: []application.Service{
			application.NewService(dataService),
			application.NewService(orderService),
			application.NewService(receiptService),
			application.NewService(managementService),
			application.NewService(webOrderService),
			application.NewService(authService),
			application.NewService(reportService),
			application.NewService(configStore),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// Create main window
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "Coffeeshop POS",
		Width:  1280,
		Height: 800,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
	})

	// Run the application
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
