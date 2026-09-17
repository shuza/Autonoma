package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shuza/Autonoma/internal/platform/config"
)

func TestHealthEndpoint(t *testing.T) {
	server := New(config.Config{Host: "172.0.0.1", Port: "8080"})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	server.Handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected content type %s, got %s", "application/json", contentType)
	}

	var body map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected valid json response, got error: %v", err)
	}

	if body["status"] != "ok" {
		t.Errorf("Expected status %s, got %s", "ok", body["status"])
	}
}

func TestUnknownRouteReturnsNotFound(t *testing.T) {
	server := New(config.Config{Host: "172.0.0.1", Port: "8080"})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/missing", nil)

	server.Handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("expected status code %d, got %d", http.StatusNotFound, recorder.Code)
	}
}
