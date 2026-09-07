package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	RedisAddr     string
	RedisPassword string
	WorkerCount   int
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: failed to load .env: %v", err)
	}

	workerCount, err := strconv.Atoi(getEnv("WORKER_COUNT", "4"))
	if err != nil {
		return nil, fmt.Errorf("WORKER_COUNT must be an integer")
	}
	if workerCount < 1 {
		return nil, fmt.Errorf("WORKER_COUNT must be greater than 0")
	}
	cfg := &Config{
		Port:          getEnv("PORT", "8082"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		WorkerCount:   workerCount,
	}
	if cfg.RedisAddr == "" {
		return nil, fmt.Errorf("REDIS_ADDR is required")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
