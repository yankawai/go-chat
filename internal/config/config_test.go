package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("STATIC_DIR", "")
	t.Setenv("WS_ALLOWED_ORIGINS", "")
	t.Setenv("WS_READ_LIMIT_BYTES", "")
	t.Setenv("WS_SEND_QUEUE_SIZE", "")
	t.Setenv("WS_PING_PERIOD", "")
	t.Setenv("WS_PONG_WAIT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTPAddr != defaultHTTPAddr {
		t.Fatalf("HTTPAddr = %q, want %q", cfg.HTTPAddr, defaultHTTPAddr)
	}
	if cfg.StaticDir != defaultStaticDir {
		t.Fatalf("StaticDir = %q, want %q", cfg.StaticDir, defaultStaticDir)
	}
	if cfg.WebSocket.ReadLimit != defaultReadLimit {
		t.Fatalf("ReadLimit = %d, want %d", cfg.WebSocket.ReadLimit, defaultReadLimit)
	}
}

func TestLoadValidatesWebSocketTimers(t *testing.T) {
	t.Setenv("WS_PING_PERIOD", "60s")
	t.Setenv("WS_PONG_WAIT", "60s")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want validation error")
	}
}

func TestLoadRejectsInvalidEnvironmentValues(t *testing.T) {
	t.Setenv("HTTP_READ_TIMEOUT", "slow")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want invalid duration error")
	}
}

func TestLoadParsesCSVAndDurations(t *testing.T) {
	t.Setenv("WS_ALLOWED_ORIGINS", "https://chat.example.com, http://localhost:8080")
	t.Setenv("HTTP_READ_TIMEOUT", "3s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got, want := len(cfg.WebSocket.AllowedOrigins), 2; got != want {
		t.Fatalf("len(AllowedOrigins) = %d, want %d", got, want)
	}
	if cfg.HTTP.ReadTimeout != 3*time.Second {
		t.Fatalf("ReadTimeout = %s, want 3s", cfg.HTTP.ReadTimeout)
	}
}
