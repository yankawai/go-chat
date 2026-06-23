package websocket

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/yankawai/go-chat/internal/chat"
)

const (
	defaultReadLimit     = int64(4096)
	defaultSendQueueSize = 32
	defaultMessageLimit  = 20
	defaultMessageWindow = 10 * time.Second
)

type HandlerConfig struct {
	AllowedOrigins []string
	ReadLimit      int64
	SendQueueSize  int
	WriteWait      time.Duration
	PongWait       time.Duration
	PingPeriod     time.Duration
	MessageLimit   int
	MessageWindow  time.Duration
}

type Handler struct {
	cfg      HandlerConfig
	upgrader websocket.Upgrader
	service  *chat.Service
	room     *chat.Room
	history  *chat.History
	logger   *slog.Logger
}

func NewHandler(cfg HandlerConfig, service *chat.Service, room *chat.Room, history *chat.History, logger *slog.Logger) *Handler {
	if cfg.ReadLimit <= 0 {
		cfg.ReadLimit = defaultReadLimit
	}
	if cfg.SendQueueSize <= 0 {
		cfg.SendQueueSize = defaultSendQueueSize
	}
	if cfg.WriteWait <= 0 {
		cfg.WriteWait = 10 * time.Second
	}
	if cfg.PongWait <= 0 {
		cfg.PongWait = 60 * time.Second
	}
	if cfg.PingPeriod <= 0 {
		cfg.PingPeriod = (cfg.PongWait * 9) / 10
	}
	if cfg.MessageLimit <= 0 {
		cfg.MessageLimit = defaultMessageLimit
	}
	if cfg.MessageWindow <= 0 {
		cfg.MessageWindow = defaultMessageWindow
	}
	if logger == nil {
		logger = slog.Default()
	}

	handler := &Handler{
		cfg:     cfg,
		service: service,
		room:    room,
		history: history,
		logger:  logger,
	}
	handler.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     handler.checkOrigin,
	}

	return handler
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.service == nil || h.room == nil {
		http.Error(w, "websocket dependencies are not configured", http.StatusInternalServerError)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Warn("upgrade websocket", "error", err)
		return
	}

	clientID := uuid.NewString()
	client := newClient(clientID, conn, h.cfg.SendQueueSize, h.cfg.WriteWait, h.logger)
	if err := h.room.Register(client); err != nil {
		h.logger.Warn("register websocket client", "client_id", clientID, "error", err)
		_ = client.Close()
		return
	}

	go func() {
		client.writeLoop(h.cfg.PingPeriod)
		h.room.Unregister(client.ID())
	}()

	limiter := newRateLimiter(h.cfg.MessageLimit, h.cfg.MessageWindow, nil)
	h.readLoop(r.Context(), client, limiter)
	h.room.Unregister(client.ID())
}

func (h *Handler) readLoop(ctx context.Context, client *client, limiter *rateLimiter) {
	defer h.logger.Info("client reader stopped", "client_id", client.ID())

	client.conn.SetReadLimit(h.cfg.ReadLimit)
	if err := client.conn.SetReadDeadline(time.Now().Add(h.cfg.PongWait)); err != nil {
		h.logger.Warn("set read deadline", "client_id", client.ID(), "error", err)
		return
	}
	client.conn.SetPongHandler(func(string) error {
		return client.conn.SetReadDeadline(time.Now().Add(h.cfg.PongWait))
	})

	for {
		var inbound inboundEvent
		if err := client.conn.ReadJSON(&inbound); err != nil {
			if !isExpectedClose(err) {
				h.logger.Warn("read websocket message", "client_id", client.ID(), "error", err)
			}
			return
		}
		if !limiter.Allow() {
			notice := h.service.SystemNotice("message rejected: rate limit exceeded")
			if sendErr := client.Send(ctx, notice); sendErr != nil {
				h.logger.Warn("send rate limit notice", "client_id", client.ID(), "error", sendErr)
				return
			}
			continue
		}

		event, err := h.service.NewMessage(chat.MessageInput{
			User:  inbound.User,
			Color: inbound.Color,
			Text:  inbound.text(),
		})
		if err != nil {
			notice := h.service.SystemNotice("message rejected: " + err.Error())
			if sendErr := client.Send(ctx, notice); sendErr != nil {
				h.logger.Warn("send validation notice", "client_id", client.ID(), "error", sendErr)
				return
			}
			continue
		}

		h.room.Broadcast(ctx, event)
		if h.history != nil {
			h.history.Append(event)
		}
	}
}

func (h *Handler) checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}

	if len(h.cfg.AllowedOrigins) > 0 {
		return originAllowed(origin, h.cfg.AllowedOrigins)
	}

	host := r.Host
	return sameHostOrigin(origin, host) || isLoopbackOrigin(origin)
}

func isExpectedClose(err error) bool {
	return websocket.IsCloseError(
		err,
		websocket.CloseGoingAway,
		websocket.CloseNormalClosure,
		websocket.CloseNoStatusReceived,
	) || errors.Is(err, context.Canceled)
}

type inboundEvent struct {
	User  string `json:"user"`
	Color string `json:"color"`
	Msg   string `json:"msg"`
	Text  string `json:"text"`
}

func (e inboundEvent) text() string {
	if strings.TrimSpace(e.Text) != "" {
		return e.Text
	}
	return e.Msg
}

type outboundEvent struct {
	ID        string `json:"id"`
	Sequence  uint64 `json:"sequence"`
	Type      string `json:"type"`
	User      string `json:"user"`
	Color     string `json:"color,omitempty"`
	Msg       string `json:"msg"`
	Text      string `json:"text"`
	CreatedAt string `json:"createdAt"`
}

func toOutboundEvent(event chat.Event) outboundEvent {
	return outboundEvent{
		ID:        event.ID,
		Sequence:  event.Sequence,
		Type:      string(event.Type),
		User:      event.User,
		Color:     event.Color,
		Msg:       event.Text,
		Text:      event.Text,
		CreatedAt: event.CreatedAt.Format(time.RFC3339Nano),
	}
}
