package websocket

import (
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yankawai/go-chat/internal/chat"
)

func TestHandlerBroadcastsMessages(t *testing.T) {
	now := time.Date(2026, 6, 23, 10, 0, 0, 0, time.UTC)
	service := chat.NewService(chat.ServiceConfig{
		Now: func() time.Time { return now },
		NewID: func() string {
			return "message-id"
		},
	})
	room := chat.NewRoom(slog.Default())
	history := chat.NewHistory(10)
	handler := NewHandler(HandlerConfig{
		PingPeriod:    time.Second,
		PongWait:      2 * time.Second,
		WriteWait:     time.Second,
		SendQueueSize: 4,
		ReadLimit:     1024,
	}, service, room, history, slog.Default())

	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	first := dialWebSocket(t, wsURL)
	defer first.Close()
	second := dialWebSocket(t, wsURL)
	defer second.Close()

	if err := first.WriteJSON(inboundEvent{
		User:  " yan ",
		Color: "#AA00FF",
		Msg:   " hello ",
	}); err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}

	firstEvent := readOutboundEvent(t, first)
	secondEvent := readOutboundEvent(t, second)

	for _, event := range []outboundEvent{firstEvent, secondEvent} {
		if event.ID != "message-id" {
			t.Fatalf("ID = %q, want message-id", event.ID)
		}
		if event.Sequence != 1 {
			t.Fatalf("Sequence = %d, want 1", event.Sequence)
		}
		if event.Type != string(chat.EventTypeMessage) {
			t.Fatalf("Type = %q, want %q", event.Type, chat.EventTypeMessage)
		}
		if event.User != "yan" {
			t.Fatalf("User = %q, want yan", event.User)
		}
		if event.Color != "#aa00ff" {
			t.Fatalf("Color = %q, want #aa00ff", event.Color)
		}
		if event.Msg != "hello" || event.Text != "hello" {
			t.Fatalf("message text = msg:%q text:%q, want hello", event.Msg, event.Text)
		}
		if event.CreatedAt != now.Format(time.RFC3339Nano) {
			t.Fatalf("CreatedAt = %q, want %q", event.CreatedAt, now.Format(time.RFC3339Nano))
		}
	}

	if history.Len() != 1 {
		t.Fatalf("history len = %d, want 1", history.Len())
	}
}

func dialWebSocket(t *testing.T, url string) *websocket.Conn {
	t.Helper()

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}

	return conn
}

func readOutboundEvent(t *testing.T, conn *websocket.Conn) outboundEvent {
	t.Helper()

	if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("SetReadDeadline() error = %v", err)
	}

	var event outboundEvent
	if err := conn.ReadJSON(&event); err != nil {
		t.Fatalf("ReadJSON() error = %v", err)
	}

	return event
}
