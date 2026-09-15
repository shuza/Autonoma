package httpserver

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/shuza/Autonoma/internal/platform/config"
	"github.com/shuza/Autonoma/internal/platform/health"
)

func New(cfg config.Config) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handleHealth)
	mux.HandleFunc("/", handleNotFound)
	return &http.Server{
		Addr:              cfg.Address(),
		Handler:           mux,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(health.Check()); err != nil {
		slog.Error("failed health check", err)
		http.Error(w, "failed health check", http.StatusInternalServerError)
	}
}

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}
