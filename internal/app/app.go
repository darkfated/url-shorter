package app

import (
	"fmt"
	"net/http"

	"url-shorter/internal/config"
	"url-shorter/internal/handler"
	"url-shorter/internal/service"
	"url-shorter/internal/storage/memory"
	"url-shorter/internal/storage/postgresql"
)

func Run() error {
	cfg := config.Load()

	store, err := newStore(cfg)
	if err != nil {
		return err
	}
	if closer, ok := store.(interface{ Close() error }); ok {
		defer func() {
			_ = closer.Close()
		}()
	}

	svc := service.New(store)
	h := handler.New(svc, cfg.PublicBaseURL)

	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: h.Routes(),
	}

	return server.ListenAndServe()
}

func newStore(cfg config.Config) (service.Store, error) {
	switch cfg.StorageType {
	case config.StorageMemory:
		return memory.New(), nil
	case config.StoragePostgreSQL:
		return postgresql.New(cfg.PostgreSQLDSN)
	default:
		return nil, fmt.Errorf("неизвестный тип хранилища: %s", cfg.StorageType)
	}
}
