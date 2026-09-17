package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/shuza/Autonoma/internal/app"
	"github.com/shuza/Autonoma/internal/httpserver"
	"github.com/shuza/Autonoma/internal/platform/config"
)

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application, err := app.New(context.Background(), cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer application.Close()

	server := httpserver.NewWithDependencies(cfg, application.LeadService)

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("starting api server on %s", cfg.Port)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatal(err)
		}
	}
}
