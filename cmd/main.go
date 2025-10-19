package main

import (
	"log"
	"os"

	"main-service/config"
	"main-service/internal/db"
	httpHandler "main-service/internal/delivery/http"
	"main-service/internal/middleware"
	"main-service/internal/repository"
	"main-service/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// ===== Load konfigurasi dari .env =====
	cfg := config.Load()

	// ===== Koneksi ke database =====
	dbConn, err := db.NewMySQLConnection(cfg)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// ===== Repository =====
	userRepo := repository.NewUserRepository(dbConn)
	featureRepo := repository.NewFeatureRepository(dbConn)
	matpelRepo := repository.NewMatpelRepository(dbConn)
	bimbelRepo := repository.NewBimbelRepository(dbConn)

	// ===== Usecase =====
	userUC := usecase.NewUserUsecase(userRepo, cfg.JWTSecret, cfg.JWTExpHour)
	featureUC := usecase.NewFeatureUsecase(featureRepo)
	matpelUC := usecase.NewMatpelUsecase(matpelRepo, featureRepo)
	bimbelUC := usecase.NewBimbelUsecase(bimbelRepo)

	// ===== Handler (HTTP Delivery) =====
	userHandler := httpHandler.NewUserHandler(userUC)
	featureHandler := httpHandler.NewFeatureHandler(featureUC)
	matpelHandler := httpHandler.NewMatpelHandler(matpelUC)
	bimbelHandler := httpHandler.NewBimbelHandler(bimbelUC, userRepo)

	// ===== Fiber Setup =====
	app := fiber.New()

	// ===== Buat folder uploads jika belum ada =====
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		if err := os.MkdirAll("uploads/thumbnails", os.ModePerm); err != nil {
			log.Fatalf("Failed to create uploads folder: %v", err)
		}
	}

	// ===== Static file serving (akses: http://localhost:8080/uploads/...) =====
	app.Static("/uploads", "./uploads")

	// ===== Routes =====
	api := app.Group("/api/v1")

	// Public routes (tanpa login)
	userHandler.RegisterRoutes(api) // Login & Register

	// Protected routes (harus login)
	protected := api.Group("") // group kosong untuk endpoint di bawahnya
	protected.Use(middleware.AuthMiddleware())
	featureHandler.RegisterRoutes(protected)
	matpelHandler.RegisterRoutes(protected)
	bimbelHandler.RegisterRoutes(protected)

	// ===== Jalankan server =====
	log.Printf("🚀 Server running on port %s", cfg.AppPort)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
