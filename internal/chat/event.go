package chat

import "time"

type EventType string

const (
	EventTypeMessage EventType = "message"
	EventTypeSystem  EventType = "system"
)

type Event struct {
	ID        string
	Type      EventType
	User      string
	Color     string
	Text      string
	CreatedAt time.Time
}

type MessageInput struct {
	User  string
	Color string
	Text  string
}
