package chat

import (
	"context"
	"errors"
	"log/slog"
	"sync"
)

var (
	ErrClientAlreadyRegistered = errors.New("client already registered")
	ErrClientIDRequired        = errors.New("client id is required")
)

type Client interface {
	ID() string
	Send(context.Context, Event) error
	Close() error
}

type Room struct {
	logger  *slog.Logger
	mu      sync.RWMutex
	clients map[string]Client
	stats   RoomStats
}

type RoomStats struct {
	ActiveClients       int   `json:"activeClients"`
	TotalConnections    int64 `json:"totalConnections"`
	TotalDisconnections int64 `json:"totalDisconnections"`
	TotalBroadcasts     int64 `json:"totalBroadcasts"`
	TotalDeliveries     int64 `json:"totalDeliveries"`
}

func NewRoom(logger *slog.Logger) *Room {
	if logger == nil {
		logger = slog.Default()
	}

	return &Room{
		logger:  logger,
		clients: make(map[string]Client),
	}
}

func (r *Room) Register(client Client) error {
	if client.ID() == "" {
		return ErrClientIDRequired
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.clients[client.ID()]; exists {
		return ErrClientAlreadyRegistered
	}

	r.clients[client.ID()] = client
	r.stats.TotalConnections++
	r.logger.Info("client registered", "client_id", client.ID(), "clients", len(r.clients))
	return nil
}

func (r *Room) Unregister(clientID string) {
	r.mu.Lock()
	client, exists := r.clients[clientID]
	if exists {
		delete(r.clients, clientID)
		r.stats.TotalDisconnections++
	}
	remaining := len(r.clients)
	r.mu.Unlock()

	if !exists {
		return
	}

	if err := client.Close(); err != nil {
		r.logger.Warn("close client", "client_id", clientID, "error", err)
	}
	r.logger.Info("client unregistered", "client_id", clientID, "clients", remaining)
}

func (r *Room) Broadcast(ctx context.Context, event Event) int {
	clients := r.snapshot()
	delivered := 0

	for _, client := range clients {
		if err := client.Send(ctx, event); err != nil {
			r.logger.Warn("drop slow client", "client_id", client.ID(), "error", err)
			r.Unregister(client.ID())
			continue
		}
		delivered++
	}

	r.recordBroadcast(delivered)

	return delivered
}

func (r *Room) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients)
}

func (r *Room) Stats() RoomStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := r.stats
	stats.ActiveClients = len(r.clients)
	return stats
}

func (r *Room) snapshot() []Client {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clients := make([]Client, 0, len(r.clients))
	for _, client := range r.clients {
		clients = append(clients, client)
	}

	return clients
}

func (r *Room) recordBroadcast(delivered int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.stats.TotalBroadcasts++
	r.stats.TotalDeliveries += int64(delivered)
}
