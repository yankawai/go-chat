package chat

import (
	"errors"
	"testing"
	"time"
)

func TestServiceNewMessageNormalizesValidInput(t *testing.T) {
	now := time.Date(2026, 6, 23, 10, 0, 0, 0, time.UTC)
	service := NewService(ServiceConfig{
		Now: func() time.Time { return now },
		NewID: func() string {
			return "message-id"
		},
	})

	event, err := service.NewMessage(MessageInput{
		User:  "  yan  ",
		Color: "#AA00ff",
		Text:  " hello ",
	})
	if err != nil {
		t.Fatalf("NewMessage() error = %v", err)
	}

	if event.ID != "message-id" {
		t.Fatalf("ID = %q, want message-id", event.ID)
	}
	if event.User != "yan" {
		t.Fatalf("User = %q, want yan", event.User)
	}
	if event.Color != "#aa00ff" {
		t.Fatalf("Color = %q, want normalized hex", event.Color)
	}
	if event.Text != "hello" {
		t.Fatalf("Text = %q, want hello", event.Text)
	}
	if !event.CreatedAt.Equal(now) {
		t.Fatalf("CreatedAt = %s, want %s", event.CreatedAt, now)
	}
}

func TestServiceNewMessageDefaultsColor(t *testing.T) {
	service := NewService(ServiceConfig{NewID: func() string { return "id" }})

	event, err := service.NewMessage(MessageInput{User: "yan", Text: "hello"})
	if err != nil {
		t.Fatalf("NewMessage() error = %v", err)
	}

	if event.Color != defaultUserColor {
		t.Fatalf("Color = %q, want %q", event.Color, defaultUserColor)
	}
}

func TestServiceNewMessageRejectsInvalidInput(t *testing.T) {
	service := NewService(ServiceConfig{})

	tests := []struct {
		name  string
		input MessageInput
		err   error
	}{
		{name: "empty user", input: MessageInput{Text: "hello"}, err: ErrEmptyUser},
		{name: "empty text", input: MessageInput{User: "yan"}, err: ErrEmptyMessage},
		{name: "invalid color", input: MessageInput{User: "yan", Color: "red", Text: "hello"}, err: ErrInvalidColor},
		{name: "long user", input: MessageInput{User: "123456789012345678901", Text: "hello"}, err: ErrUserTooLong},
		{name: "long message", input: MessageInput{User: "yan", Text: stringOfRunes(MaxMessageLength + 1)}, err: ErrMessageTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.NewMessage(tt.input)
			if !errors.Is(err, tt.err) {
				t.Fatalf("NewMessage() error = %v, want %v", err, tt.err)
			}
		})
	}
}

func stringOfRunes(count int) string {
	runes := make([]rune, count)
	for i := range runes {
		runes[i] = 'a'
	}
	return string(runes)
}
