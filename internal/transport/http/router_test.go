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
	"github.com/yankawai/go-chat/internal/chat"
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

func TestReadyHandler(t *testing.T) {
	router := NewRouter(RouterConfig{}, http.NotFoundHandler(), slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
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

func TestAPINotFoundHandler(t *testing.T) {
	router := NewRouter(RouterConfig{}, http.NotFoundHandler(), slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/api/missing", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var body errorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Error.Code != "not_found" {
		t.Fatalf("error code = %q, want not_found", body.Error.Code)
	}
}

func TestConstraintsHandler(t *testing.T) {
	router := NewRouter(RouterConfig{}, http.NotFoundHandler(), slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/api/constraints", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body constraintsResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.MaxUserLength == 0 || body.MaxMessageLength == 0 || body.DefaultColor == "" {
		t.Fatalf("constraints response is incomplete: %+v", body)
	}
}

func TestRoomHandler(t *testing.T) {
	room := chat.NewRoom(slog.Default())
	router := NewRouter(RouterConfig{Room: room}, http.NotFoundHandler(), slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/api/room", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body chat.RoomStats
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.ActiveClients != 0 {
		t.Fatalf("ActiveClients = %d, want 0", body.ActiveClients)
	}
}

func TestIndexHandlerReturnsJSONWhenIndexMissing(t *testing.T) {
	router := NewRouter(RouterConfig{StaticDir: t.TempDir()}, http.NotFoundHandler(), slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var body errorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Error.Code != "index_not_available" {
		t.Fatalf("error code = %q, want index_not_available", body.Error.Code)
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
