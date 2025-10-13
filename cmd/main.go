package main

import (
	"log"

	"main-service/config"
	"main-service/internal/db"
	httpHandler "main-service/internal/delivery/http"
	"main-service/internal/middleware"
	"main-service/internal/repository"
	"main-service/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Load konfigurasi dari .env
	cfg := config.Load()

	// Koneksi ke database
	dbConn, err := db.NewMySQLConnection(cfg)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// ===== Repository =====
	userRepo := repository.NewUserRepository(dbConn)
	featureRepo := repository.NewFeatureRepository(dbConn)
	matpelRepo := repository.NewMatpelRepository(dbConn)

	// ===== Usecase =====
	userUC := usecase.NewUserUsecase(userRepo, cfg.JWTSecret, cfg.JWTExpHour)
	featureUC := usecase.NewFeatureUsecase(featureRepo)
	matpelUC := usecase.NewMatpelUsecase(matpelRepo, featureRepo)

	// ===== Handler (HTTP Delivery) =====
	userHandler := httpHandler.NewUserHandler(userUC)
	featureHandler := httpHandler.NewFeatureHandler(featureUC)
	matpelHandler := httpHandler.NewMatpelHandler(matpelUC)

	// ===== Fiber Setup =====
	app := fiber.New()

	// === ROUTES ===
	api := app.Group("/api/v1")

	// Public routes (tanpa login)
	userHandler.RegisterRoutes(api) // Login & Register

	// Protected routes (harus login)
	protected := api.Group("") // group kosong untuk endpoint di bawahnya
	protected.Use(middleware.AuthMiddleware())
	featureHandler.RegisterRoutes(protected)
	matpelHandler.RegisterRoutes(protected)

	// Jalankan server
	log.Printf("Server running on port %s", cfg.AppPort)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
