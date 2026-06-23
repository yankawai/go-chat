package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/yankawai/go-chat/internal/build"
	"github.com/yankawai/go-chat/internal/chat"
)

type RouterConfig struct {
	StaticDir string
	BuildInfo build.Info
	Room      *chat.Room
	History   *chat.History
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
	mux.HandleFunc("GET /api/constraints", constraintsHandler)
	mux.HandleFunc("GET /api/room", roomHandler(cfg.Room))
	mux.HandleFunc("GET /api/messages", messagesHandler(cfg.History))
	mux.HandleFunc("GET /api/", apiNotFoundHandler)
	mux.Handle("GET /ws", wsHandler)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(cfg.StaticDir))))
	mux.HandleFunc("GET /", indexHandler(cfg.StaticDir, logger))

	return requestID(accessLog(recoverPanic(securityHeaders(mux), logger), logger))
}

func infoHandler(info build.Info) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, info)
	}
}

func apiNotFoundHandler(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusNotFound, "not_found", "API endpoint was not found")
}

type constraintsResponse struct {
	MaxUserLength    int    `json:"maxUserLength"`
	MaxMessageLength int    `json:"maxMessageLength"`
	DefaultColor     string `json:"defaultColor"`
}

func constraintsHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, constraintsResponse{
		MaxUserLength:    chat.MaxUserLength,
		MaxMessageLength: chat.MaxMessageLength,
		DefaultColor:     chat.DefaultUserColor,
	})
}

func roomHandler(room *chat.Room) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if room == nil {
			writeError(w, http.StatusServiceUnavailable, "room_unavailable", "chat room is not available")
			return
		}
		writeJSON(w, http.StatusOK, room.Stats())
	}
}

type messageResponse struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	User      string `json:"user"`
	Color     string `json:"color,omitempty"`
	Text      string `json:"text"`
	CreatedAt string `json:"createdAt"`
}

func messagesHandler(history *chat.History) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if history == nil {
			writeError(w, http.StatusServiceUnavailable, "history_unavailable", "chat history is not available")
			return
		}

		limit, err := parsePositiveIntQuery(r, "limit")
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_limit", err.Error())
			return
		}

		events := history.List(limit)
		response := make([]messageResponse, 0, len(events))
		for _, event := range events {
			response = append(response, messageResponse{
				ID:        event.ID,
				Type:      string(event.Type),
				User:      event.User,
				Color:     event.Color,
				Text:      event.Text,
				CreatedAt: event.CreatedAt.Format(time.RFC3339Nano),
			})
		}

		writeJSON(w, http.StatusOK, response)
	}
}

func parsePositiveIntQuery(r *http.Request, key string) (int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return 0, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, errors.New(key + " must be a positive integer")
	}

	return value, nil
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
