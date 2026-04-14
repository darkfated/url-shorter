package config

import (
	"errors"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type StorageType string

const (
	StorageMemory     StorageType = "memory"
	StoragePostgreSQL StorageType = "postgres"
)

type Config struct {
	HTTPAddr      string
	PublicBaseURL string
	StorageType   StorageType
	PostgreSQLDSN string
}

func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		HTTPAddr:      envOrDefault("HTTP_ADDR", ":8080"),
		PublicBaseURL: envOrDefault("PUBLIC_BASE_URL", "http://localhost:8080"),
		StorageType:   StorageType(envOrDefault("STORAGE_TYPE", "memory")),
		PostgreSQLDSN: os.Getenv("POSTGRES_DSN"),
	}
	return cfg
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.HTTPAddr) == "" {
		return errors.New("адрес сервера не задан")
	}
	if strings.TrimSpace(c.PublicBaseURL) == "" {
		return errors.New("публичный адрес не задан")
	}
	if c.StorageType == StoragePostgreSQL && strings.TrimSpace(c.PostgreSQLDSN) == "" {
		return errors.New("dsn для postgres не задан")
	}
	return nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
