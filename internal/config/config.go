// Package config загружает конфигурацию из окружения (.env опционально).
package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config — параметры приложения.
type Config struct {
	Port        string
	Env         string
	DatabaseURL string
	BotToken    string
	BotUsername string
	JWTSecret   string

	// InitDataMaxAge — максимальный возраст initData; 0 — не проверять (dev).
	InitDataMaxAge time.Duration
	// NotifyInterval — период опроса outbox воркером.
	NotifyInterval time.Duration
}

// Load читает переменные окружения (и .env, если есть).
func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		Port:           getenv("APP_PORT", "8080"),
		Env:            getenv("APP_ENV", "dev"),
		DatabaseURL:    getenv("DATABASE_URL", "postgres://cargo:cargo@localhost:5432/cargo?sslmode=disable"),
		BotToken:       os.Getenv("BOT_TOKEN"),
		BotUsername:    getenv("BOT_USERNAME", "cargo_demo_bot"),
		JWTSecret:      getenv("JWT_SECRET", "dev-secret-change-me"),
		InitDataMaxAge: time.Duration(getenvInt("INIT_DATA_MAX_AGE_SECONDS", 0)) * time.Second,
		NotifyInterval: time.Duration(getenvInt("NOTIFY_INTERVAL_SECONDS", 5)) * time.Second,
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
