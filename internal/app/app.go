package app

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

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
		Addr:              cfg.HTTPAddr,
		Handler:           h.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	case <-ctx.Done():
		if err := server.Shutdown(context.Background()); err != nil {
			return err
		}
	}

	err = <-errCh
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
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
