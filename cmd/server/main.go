package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yankawai/go-chat/internal/build"
	"github.com/yankawai/go-chat/internal/chat"
	"github.com/yankawai/go-chat/internal/config"
	httptransport "github.com/yankawai/go-chat/internal/transport/http"
	wstransport "github.com/yankawai/go-chat/internal/transport/websocket"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))

	chatService := chat.NewService(chat.ServiceConfig{})
	room := chat.NewRoom(logger.With("component", "chat_room"))
	history := chat.NewHistory(cfg.Chat.HistoryLimit)
	wsHandler := wstransport.NewHandler(wstransport.HandlerConfig{
		AllowedOrigins: cfg.WebSocket.AllowedOrigins,
		PingPeriod:     cfg.WebSocket.PingPeriod,
		PongWait:       cfg.WebSocket.PongWait,
		ReadLimit:      cfg.WebSocket.ReadLimit,
		SendQueueSize:  cfg.WebSocket.SendQueueSize,
		WriteWait:      cfg.WebSocket.WriteWait,
	}, chatService, room, history, logger.With("component", "websocket"))

	router := httptransport.NewRouter(httptransport.RouterConfig{
		StaticDir: cfg.StaticDir,
		BuildInfo: build.NewInfo(cfg.AppName),
		Room:      room,
	}, wsHandler, logger.With("component", "http"))

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server started", "addr", cfg.HTTPAddr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown requested")
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
		return
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			logger.Error("server stopped with error", "error", err)
			os.Exit(1)
		}
	default:
	}

	logger.Info("server stopped")
}
