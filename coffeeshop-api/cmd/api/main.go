package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"coffeeshop-api/internal/config"
	"coffeeshop-api/internal/database"
	"coffeeshop-api/internal/handler"
	"coffeeshop-api/internal/middleware"
	"coffeeshop-api/internal/repository"
	"coffeeshop-api/internal/service"
	"coffeeshop-api/internal/sse"
	"coffeeshop-api/internal/storage"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Setup structured logging
	logLevel := slog.LevelInfo
	if cfg.IsDevelopment() {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("connected to database")

	// Wire dependencies: repositories
	userRepo := repository.NewUserRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	menuItemRepo := repository.NewMenuItemRepository(db)
	inventoryRepo := repository.NewInventoryRepository(db)
	recipeRepo := repository.NewRecipeRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	tableRepo := repository.NewTableRepository(db)

	// Wire dependencies: services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	categoryService := service.NewCategoryService(categoryRepo)
	menuItemService := service.NewMenuItemService(menuItemRepo, categoryRepo)
	inventoryService := service.NewInventoryService(inventoryRepo, recipeRepo, menuItemRepo)
	recipeService := service.NewRecipeService(recipeRepo, menuItemRepo, inventoryRepo)
	orderService := service.NewOrderService(orderRepo)
	tableService := service.NewTableService(tableRepo)

	// SSE hub (shared between handlers for real-time order events)
	sseHub := sse.NewHub()

	// Wire dependencies: handlers
	authHandler := handler.NewAuthHandler(authService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	menuItemHandler := handler.NewMenuItemHandler(menuItemService)
	inventoryHandler := handler.NewInventoryHandler(inventoryService)
	recipeHandler := handler.NewRecipeHandler(recipeService)
	orderHandler := handler.NewOrderHandler(orderService, sseHub)
	tableHandler := handler.NewTableHandler(tableService)
	webOrderHandler := handler.NewWebOrderHandler(orderService, tableService, sseHub)
	sseHandler := handler.NewSSEHandler(sseHub)

	// R2 storage (nil if not configured)
	r2, err := storage.NewR2Storage()
	if err != nil {
		slog.Warn("R2 storage init failed", "error", err)
	}
	if r2 != nil {
		slog.Info("R2 storage configured")
	} else {
		slog.Warn("R2 storage not configured (uploads will be unavailable)")
	}
	uploadHandler := handler.NewUploadHandler(r2)

	// Auth middleware
	authMw := middleware.Auth(cfg.JWTSecret)

	// Build router
	mux := http.NewServeMux()

	// Public auth routes
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)

	// Public read routes
	mux.HandleFunc("GET /api/v1/categories", categoryHandler.List)
	mux.HandleFunc("GET /api/v1/categories/{id}", categoryHandler.Get)
	mux.HandleFunc("GET /api/v1/menu-items", menuItemHandler.List)
	mux.HandleFunc("GET /api/v1/menu-items/{id}", menuItemHandler.Get)

	// Protected category routes
	mux.HandleFunc("POST /api/v1/categories", authMw(categoryHandler.Create))
	mux.HandleFunc("PUT /api/v1/categories/{id}", authMw(categoryHandler.Update))
	mux.HandleFunc("DELETE /api/v1/categories/{id}", authMw(categoryHandler.Delete))

	// Protected menu item routes
	mux.HandleFunc("POST /api/v1/menu-items", authMw(menuItemHandler.Create))
	mux.HandleFunc("PUT /api/v1/menu-items/{id}", authMw(menuItemHandler.Update))
	mux.HandleFunc("DELETE /api/v1/menu-items/{id}", authMw(menuItemHandler.Delete))

	// Protected inventory routes
	mux.HandleFunc("GET /api/v1/inventory", authMw(inventoryHandler.List))
	mux.HandleFunc("GET /api/v1/inventory/{id}", authMw(inventoryHandler.Get))
	mux.HandleFunc("POST /api/v1/inventory", authMw(inventoryHandler.Create))
	mux.HandleFunc("PUT /api/v1/inventory/{id}", authMw(inventoryHandler.Update))
	mux.HandleFunc("DELETE /api/v1/inventory/{id}", authMw(inventoryHandler.Delete))
	mux.HandleFunc("POST /api/v1/inventory/adjust", authMw(inventoryHandler.Adjust))

	// Protected order routes
	mux.HandleFunc("GET /api/v1/orders", authMw(orderHandler.List))
	mux.HandleFunc("POST /api/v1/orders", authMw(orderHandler.Create))
	mux.HandleFunc("PUT /api/v1/orders/{id}/status", authMw(orderHandler.UpdateStatus))

	// Protected recipe routes
	mux.HandleFunc("GET /api/v1/menu-items/{id}/recipe", authMw(recipeHandler.Get))
	mux.HandleFunc("PUT /api/v1/menu-items/{id}/recipe", authMw(recipeHandler.Set))

	// Protected table routes
	mux.HandleFunc("GET /api/v1/tables", authMw(tableHandler.List))
	mux.HandleFunc("POST /api/v1/tables", authMw(tableHandler.Create))
	mux.HandleFunc("DELETE /api/v1/tables/{id}", authMw(tableHandler.Delete))

	// Web order (public, table-token auth via query param)
	mux.HandleFunc("POST /api/v1/web-orders", webOrderHandler.Create)

	// SSE stream (auth required)
	mux.HandleFunc("GET /api/v1/orders/stream", authMw(sseHandler.Stream))

	// File upload (auth required)
	mux.HandleFunc("POST /api/v1/uploads", authMw(uploadHandler.Upload))

	// Apply global middleware
	finalHandler := middleware.Chain(mux,
		middleware.Logger,
		middleware.Recoverer,
		middleware.CORS,
	)

	// Create server (no ReadTimeout/WriteTimeout to support long-lived SSE connections)
	srv := &http.Server{
		Addr:        ":" + cfg.Port,
		Handler:     finalHandler,
		IdleTimeout: 60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		slog.Info("server starting", "port", cfg.Port, "env", cfg.Environment)
		fmt.Printf("\n  🚀 Coffeeshop API running at http://localhost:%s\n\n", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped gracefully")
}
