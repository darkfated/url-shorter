package config

import "testing"

func TestValidateMemoryConfig(t *testing.T) {
	cfg := Config{
		HTTPAddr:      ":8080",
		PublicBaseURL: "http://localhost:8080",
		StorageType:   StorageMemory,
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestValidatePostgresConfig(t *testing.T) {
	cfg := Config{
		HTTPAddr:      ":8080",
		PublicBaseURL: "http://localhost:8080",
		StorageType:   StoragePostgreSQL,
		PostgreSQLDSN: "postgres://postgres:password@localhost:5432/url_shorter?sslmode=disable",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestValidatePostgresConfigWithoutDSN(t *testing.T) {
	cfg := Config{
		HTTPAddr:      ":8080",
		PublicBaseURL: "http://localhost:8080",
		StorageType:   StoragePostgreSQL,
	}

	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected error")
	}
}
