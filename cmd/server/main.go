package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/pickIQ/oredis/internal/config"
	"github.com/pickIQ/oredis/internal/httpapi"
	"github.com/pickIQ/oredis/internal/middleware"
	"github.com/pickIQ/oredis/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	st := store.NewInMemoryStore()
	mux := httpapi.NewRouter(st)

	addr := fmt.Sprintf(":%d", cfg.Port)
	slog.Info("server starting on", "port", addr)

	if err := http.ListenAndServe(addr, middleware.Logging(middleware.Recovery(mux))); err != nil {
		log.Fatal(err)
	}
}
