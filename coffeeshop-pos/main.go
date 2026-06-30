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
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
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

	// Create sync client and worker
	apiClient := posSync.NewAPIClient(cfg.APIBaseURL)
	syncWorker := posSync.NewWorker(apiClient, db)

	// ConfigStore: persistent API connection (must be created before services that depend on it)
	configStore := service.NewConfigStoreService(db, apiClient)

	// Try auto-login using stored credentials (from previous setup)
	configStore.TryAutoLogin()

	// Create services
	versionService := service.NewVersionService()
	updateService := service.NewUpdateService(configStore)
	dataService := service.NewDataService(db, configStore, cfg.ImageCacheDir)
	syncWorker.OnMenuPulled = dataService.PreCacheImages
	orderService := service.NewOrderService(db, configStore)
	receiptService := service.NewReceiptService("المقهى")
	authService := service.NewAuthService(db)
	reportService := service.NewReportService(db)

	// Seed default admin if no users exist (first launch)
	authService.SeedDefaultAdmin()

	// Management service (requires apiClient + syncWorker)
	managementService := service.NewManagementService(db, apiClient, syncWorker)

	// Web order service (manages incoming web menu orders)
	webOrderService := service.NewWebOrderService(db, apiClient, configStore)

	// Sync dashboard service (exposes sync health to frontend)
	syncService := service.NewSyncService(syncWorker, db)

	// SSE client for real-time cross-POS sync
	sseClient := posSync.NewSSEClient(apiClient, func(event posSync.SSEEvent) {
		// Forward web order events to WebOrderService (existing behavior)
		webOrderService.HandleSSEEvent(event)

		// Trigger immediate sync on relevant events
		switch event.Type {
		case "new_order", "order_status", "data_changed":
			syncWorker.TriggerPull()
		}
	})

	// Start sync worker and SSE client in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go syncWorker.Start(ctx, cfg.SyncInterval)
	go sseClient.Connect(ctx)

	// Create Wails application
	app := application.New(application.Options{
		Name:        "Cashier POS",
		Description: "نقطة البيع - كاشير",
		Services: []application.Service{
			application.NewService(dataService),
			application.NewService(orderService),
			application.NewService(receiptService),
			application.NewService(managementService),
			application.NewService(webOrderService),
			application.NewService(authService),
			application.NewService(reportService),
			application.NewService(configStore),
			application.NewService(syncService),
			application.NewService(versionService),
			application.NewService(updateService),
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
		Title:  "Cashier POS",
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
