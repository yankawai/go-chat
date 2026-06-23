package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/yankawai/go-chat/internal/build"
)

type RouterConfig struct {
	StaticDir string
	BuildInfo build.Info
}

func NewRouter(cfg RouterConfig, wsHandler http.Handler, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	if cfg.StaticDir == "" {
		cfg.StaticDir = "static"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("GET /readyz", readyHandler)
	mux.HandleFunc("GET /api/info", infoHandler(cfg.BuildInfo))
	mux.HandleFunc("GET /api/", apiNotFoundHandler)
	mux.Handle("GET /ws", wsHandler)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(cfg.StaticDir))))
	mux.HandleFunc("GET /", indexHandler(cfg.StaticDir, logger))

	return securityHeaders(mux)
}

func infoHandler(info build.Info) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, info)
	}
}

func apiNotFoundHandler(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusNotFound, "not_found", "API endpoint was not found")
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func readyHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func indexHandler(staticDir string, logger *slog.Logger) http.HandlerFunc {
	indexPath := filepath.Join(staticDir, "index.html")

	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		if _, err := os.Stat(indexPath); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				logger.Error("stat index", "path", indexPath, "error", err)
			}
			writeError(w, http.StatusNotFound, "index_not_available", "index page is not available")
			return
		}

		http.ServeFile(w, r, indexPath)
	}
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
