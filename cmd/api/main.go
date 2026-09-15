package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/shuza/Autonoma/internal/httpserver"
	"github.com/shuza/Autonoma/internal/platform/config"
)

func main() {
	cfg := config.Load()
	server := httpserver.New(cfg)
	log.Printf("starting api server on %s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("failed to start api server: %v", err)
	}
}
