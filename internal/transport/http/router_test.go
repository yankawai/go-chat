package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/yankawai/go-chat/internal/build"
)

func TestHealthHandler(t *testing.T) {
	router := NewRouter(RouterConfig{}, http.NotFoundHandler(), slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q, want nosniff", got)
	}
}

func TestInfoHandler(t *testing.T) {
	router := NewRouter(RouterConfig{
		BuildInfo: build.Info{Service: "go-chat", Version: "test"},
	}, http.NotFoundHandler(), slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/api/info", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body build.Info
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Service != "go-chat" {
		t.Fatalf("Service = %q, want go-chat", body.Service)
	}
}

func TestIndexHandlerServesStaticIndex(t *testing.T) {
	staticDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(staticDir, "index.html"), []byte("<html>ok</html>"), 0o600); err != nil {
		t.Fatalf("write index: %v", err)
	}

	router := NewRouter(RouterConfig{StaticDir: staticDir}, http.NotFoundHandler(), slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
