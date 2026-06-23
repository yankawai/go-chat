package chat

import (
	"context"
	"errors"
	"log/slog"
	"testing"
)

func TestRoomRegisterBroadcastAndUnregister(t *testing.T) {
	room := NewRoom(slog.Default())
	client := newFakeClient("client-1")

	if err := room.Register(client); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	delivered := room.Broadcast(context.Background(), Event{ID: "message-1", Text: "hello"})
	if delivered != 1 {
		t.Fatalf("Broadcast() delivered = %d, want 1", delivered)
	}
	if got := len(client.events); got != 1 {
		t.Fatalf("client events = %d, want 1", got)
	}

	room.Unregister(client.ID())
	if !client.closed {
		t.Fatal("client was not closed")
	}
	if count := room.Count(); count != 0 {
		t.Fatalf("Count() = %d, want 0", count)
	}
}

func TestRoomStatsReportsActiveClients(t *testing.T) {
	room := NewRoom(slog.Default())
	if err := room.Register(newFakeClient("client-1")); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	stats := room.Stats()
	if stats.ActiveClients != 1 {
		t.Fatalf("ActiveClients = %d, want 1", stats.ActiveClients)
	}
	if stats.TotalConnections != 1 {
		t.Fatalf("TotalConnections = %d, want 1", stats.TotalConnections)
	}
}

func TestRoomStatsTracksBroadcastsAndDisconnections(t *testing.T) {
	room := NewRoom(slog.Default())
	if err := room.Register(newFakeClient("client-1")); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	delivered := room.Broadcast(context.Background(), Event{ID: "event-1"})
	if delivered != 1 {
		t.Fatalf("Broadcast() delivered = %d, want 1", delivered)
	}
	room.Unregister("client-1")

	stats := room.Stats()
	if stats.TotalBroadcasts != 1 {
		t.Fatalf("TotalBroadcasts = %d, want 1", stats.TotalBroadcasts)
	}
	if stats.TotalDeliveries != 1 {
		t.Fatalf("TotalDeliveries = %d, want 1", stats.TotalDeliveries)
	}
	if stats.TotalDisconnections != 1 {
		t.Fatalf("TotalDisconnections = %d, want 1", stats.TotalDisconnections)
	}
}

func TestRoomRejectsDuplicateClient(t *testing.T) {
	room := NewRoom(slog.Default())

	if err := room.Register(newFakeClient("client-1")); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	err := room.Register(newFakeClient("client-1"))
	if !errors.Is(err, ErrClientAlreadyRegistered) {
		t.Fatalf("Register() error = %v, want %v", err, ErrClientAlreadyRegistered)
	}
}

func TestRoomRejectsClientsAboveCapacity(t *testing.T) {
	room := NewRoomWithConfig(RoomConfig{MaxClients: 1}, slog.Default())
	if err := room.Register(newFakeClient("client-1")); err != nil {
		t.Fatalf("Register() first error = %v", err)
	}

	err := room.Register(newFakeClient("client-2"))
	if !errors.Is(err, ErrRoomFull) {
		t.Fatalf("Register() error = %v, want %v", err, ErrRoomFull)
	}
}

func TestRoomDropsClientWhenSendFails(t *testing.T) {
	room := NewRoom(slog.Default())
	client := newFakeClient("client-1")
	client.sendErr = errors.New("queue full")

	if err := room.Register(client); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	delivered := room.Broadcast(context.Background(), Event{ID: "message-1"})
	if delivered != 0 {
		t.Fatalf("Broadcast() delivered = %d, want 0", delivered)
	}
	if count := room.Count(); count != 0 {
		t.Fatalf("Count() = %d, want 0", count)
	}
	if !client.closed {
		t.Fatal("failed client was not closed")
	}
}

type fakeClient struct {
	id      string
	events  []Event
	sendErr error
	closed  bool
}

func newFakeClient(id string) *fakeClient {
	return &fakeClient{id: id}
}

func (c *fakeClient) ID() string {
	return c.id
}

func (c *fakeClient) Send(_ context.Context, event Event) error {
	if c.sendErr != nil {
		return c.sendErr
	}
	c.events = append(c.events, event)
	return nil
}

func (c *fakeClient) Close() error {
	c.closed = true
	return nil
}
