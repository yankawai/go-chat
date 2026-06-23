package websocket

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yankawai/go-chat/internal/chat"
)

var (
	ErrClientClosed       = errors.New("websocket client is closed")
	ErrClientBackpressure = errors.New("websocket client send queue is full")
)

type client struct {
	id        string
	conn      *websocket.Conn
	send      chan chat.Event
	writeWait time.Duration
	logger    *slog.Logger

	mu     sync.RWMutex
	closed bool
}

func newClient(id string, conn *websocket.Conn, sendQueueSize int, writeWait time.Duration, logger *slog.Logger) *client {
	return &client{
		id:        id,
		conn:      conn,
		send:      make(chan chat.Event, sendQueueSize),
		writeWait: writeWait,
		logger:    logger,
	}
}

func (c *client) ID() string {
	return c.id
}

func (c *client) Send(ctx context.Context, event chat.Event) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrClientClosed
	}

	select {
	case c.send <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return ErrClientBackpressure
	}
}

func (c *client) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	close(c.send)
	c.mu.Unlock()

	return c.conn.Close()
}

func (c *client) writeLoop(pingPeriod time.Duration) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case event, ok := <-c.send:
			if !ok {
				return
			}
			if err := c.writeJSON(event); err != nil {
				c.logger.Warn("write websocket message", "client_id", c.id, "error", err)
				return
			}
		case <-ticker.C:
			if err := c.writeControl(websocket.PingMessage, nil); err != nil {
				c.logger.Warn("write websocket ping", "client_id", c.id, "error", err)
				return
			}
		}
	}
}

func (c *client) writeJSON(event chat.Event) error {
	if err := c.conn.SetWriteDeadline(time.Now().Add(c.writeWait)); err != nil {
		return err
	}
	return c.conn.WriteJSON(toOutboundEvent(event))
}

func (c *client) writeControl(messageType int, data []byte) error {
	deadline := time.Now().Add(c.writeWait)
	return c.conn.WriteControl(messageType, data, deadline)
}
