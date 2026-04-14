package config

import (
	"os"

	"github.com/joho/godotenv"
)

type StorageType string

const (
	StorageMemory     StorageType = "memory"
	StoragePostgreSQL StorageType = "postgres"
)

type Config struct {
	HTTPAddr      string
	StorageType   StorageType
	PostgreSQLDSN string
}

func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		HTTPAddr:      envOrDefault("HTTP_ADDR", ":8080"),
		StorageType:   StorageType(envOrDefault("STORAGE_TYPE", "memory")),
		PostgreSQLDSN: os.Getenv("POSTGRES_DSN"),
	}
	return cfg
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
