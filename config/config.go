package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort    string
	DBUser     string
	DBPass     string
	DBHost     string
	DBPort     string
	DBName     string
	JWTSecret  string
	JWTExpHour int
}

func Load() *Config {
	_ = godotenv.Load(".env")

	expHour, err := strconv.Atoi(os.Getenv("JWT_EXP_HOURS"))
	if err != nil || expHour <= 0 {
		expHour = 24 // default
	}

	cfg := &Config{
		AppPort:    os.Getenv("APP_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPass:     os.Getenv("DB_PASS"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBName:     os.Getenv("DB_NAME"),
		JWTSecret:  os.Getenv("JWT_SECRET"),
		JWTExpHour: expHour,
	}

	if cfg.AppPort == "" {
		log.Fatal("APP_PORT is not set in .env")
	}

	return cfg
}
