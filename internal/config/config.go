package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHTTPAddr      = ":8080"
	defaultStaticDir     = "static"
	defaultReadLimit     = int64(4096)
	defaultSendQueueSize = 32
)

type Config struct {
	AppName   string
	HTTPAddr  string
	StaticDir string
	LogLevel  slog.Level
	HTTP      HTTPConfig
	WebSocket WebSocketConfig
	Chat      ChatConfig
}

type HTTPConfig struct {
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

type WebSocketConfig struct {
	AllowedOrigins []string
	ReadLimit      int64
	SendQueueSize  int
	WriteWait      time.Duration
	PongWait       time.Duration
	PingPeriod     time.Duration
}

type ChatConfig struct {
	HistoryLimit int
	BannedTerms  []string
}

func Load() (Config, error) {
	readHeaderTimeout, err := getDuration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}
	readTimeout, err := getDuration("HTTP_READ_TIMEOUT", 15*time.Second)
	if err != nil {
		return Config{}, err
	}
	writeTimeout, err := getDuration("HTTP_WRITE_TIMEOUT", 15*time.Second)
	if err != nil {
		return Config{}, err
	}
	idleTimeout, err := getDuration("HTTP_IDLE_TIMEOUT", 60*time.Second)
	if err != nil {
		return Config{}, err
	}
	readLimit, err := getInt64("WS_READ_LIMIT_BYTES", defaultReadLimit)
	if err != nil {
		return Config{}, err
	}
	sendQueueSize, err := getInt("WS_SEND_QUEUE_SIZE", defaultSendQueueSize)
	if err != nil {
		return Config{}, err
	}
	writeWait, err := getDuration("WS_WRITE_WAIT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	pongWait, err := getDuration("WS_PONG_WAIT", 60*time.Second)
	if err != nil {
		return Config{}, err
	}
	pingPeriod, err := getDuration("WS_PING_PERIOD", 54*time.Second)
	if err != nil {
		return Config{}, err
	}
	historyLimit, err := getInt("CHAT_HISTORY_LIMIT", 100)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		AppName:   getString("APP_NAME", "go-chat"),
		HTTPAddr:  getString("HTTP_ADDR", defaultHTTPAddr),
		StaticDir: getString("STATIC_DIR", defaultStaticDir),
		LogLevel:  parseLogLevel(getString("LOG_LEVEL", "info")),
		HTTP: HTTPConfig{
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
		},
		WebSocket: WebSocketConfig{
			AllowedOrigins: splitCSV(os.Getenv("WS_ALLOWED_ORIGINS")),
			ReadLimit:      readLimit,
			SendQueueSize:  sendQueueSize,
			WriteWait:      writeWait,
			PongWait:       pongWait,
			PingPeriod:     pingPeriod,
		},
		Chat: ChatConfig{
			HistoryLimit: historyLimit,
			BannedTerms:  splitCSV(os.Getenv("CHAT_BANNED_TERMS")),
		},
	}

	if cfg.WebSocket.ReadLimit <= 0 {
		return Config{}, fmt.Errorf("WS_READ_LIMIT_BYTES must be positive")
	}
	if cfg.WebSocket.SendQueueSize <= 0 {
		return Config{}, fmt.Errorf("WS_SEND_QUEUE_SIZE must be positive")
	}
	if cfg.WebSocket.PingPeriod >= cfg.WebSocket.PongWait {
		return Config{}, fmt.Errorf("WS_PING_PERIOD must be lower than WS_PONG_WAIT")
	}
	if cfg.Chat.HistoryLimit <= 0 {
		return Config{}, fmt.Errorf("CHAT_HISTORY_LIMIT must be positive")
	}

	return cfg, nil
}

func getString(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getDuration(key string, fallback time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}

	return value, nil
}

func getInt(key string, fallback int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}

	return value, nil
}

func getInt64(key string, fallback int64) (int64, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}

	return value, nil
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}

func parseLogLevel(raw string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
