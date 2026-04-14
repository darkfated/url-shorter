package app

import (
	"net/http"

	"url-shorter/internal/config"
	"url-shorter/internal/handler"
)

func Run() error {
	cfg := config.Load()
	_ = cfg
	h := handler.New()

	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: h.Routes(),
	}

	return server.ListenAndServe()
}
