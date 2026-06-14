package main

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken       string
	GeminiAPIKey   string
	KlingAPIKey    string
	Port           string
	WorkerPoolSize int
	DatabasePath   string
}

func LoadConfig() *Config {
	// Load .env file if present (ignore error if not found)
	_ = godotenv.Load()

	port := getEnv("PORT", "3000")
	workerPoolSize := 4
	if v := os.Getenv("WORKER_POOL_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			workerPoolSize = n
		}
	}

	databasePath := getEnv("DATABASE_PATH", "./data/prompts.db")

	cfg := &Config{
		BotToken:       os.Getenv("BOT_TOKEN"),
		GeminiAPIKey:   os.Getenv("GEMINI_API_KEY"),
		KlingAPIKey:    os.Getenv("KLING_API_KEY"),
		Port:           port,
		WorkerPoolSize: workerPoolSize,
		DatabasePath:   databasePath,
	}

	if cfg.BotToken == "" {
		slog.Error("BOT_TOKEN is required")
		os.Exit(1)
	}
	if cfg.GeminiAPIKey == "" {
		slog.Error("GEMINI_API_KEY is required")
		os.Exit(1)
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
