package chat

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

const (
	MaxUserLength    = 20
	MaxMessageLength = 1000
	DefaultUserColor = "#111827"
	systemUser       = "system"
	systemColor      = "#64748b"
)

var (
	ErrEmptyUser      = errors.New("user is required")
	ErrEmptyMessage   = errors.New("message is required")
	ErrUserTooLong    = errors.New("user is too long")
	ErrMessageTooLong = errors.New("message is too long")
	ErrInvalidColor   = errors.New("color must be a hex color")

	hexColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
)

type ServiceConfig struct {
	Now   func() time.Time
	NewID func() string
}

type Service struct {
	now      func() time.Time
	newID    func() string
	sequence atomic.Uint64
}

func NewService(cfg ServiceConfig) *Service {
	now := cfg.Now
	if now == nil {
		now = time.Now
	}

	newID := cfg.NewID
	if newID == nil {
		newID = func() string {
			return uuid.NewString()
		}
	}

	return &Service{
		now:   now,
		newID: newID,
	}
}

func (s *Service) NewMessage(input MessageInput) (Event, error) {
	user := strings.TrimSpace(input.User)
	text := strings.TrimSpace(input.Text)
	color := strings.TrimSpace(input.Color)

	switch {
	case user == "":
		return Event{}, ErrEmptyUser
	case text == "":
		return Event{}, ErrEmptyMessage
	case runeCount(user) > MaxUserLength:
		return Event{}, fmt.Errorf("%w: max %d characters", ErrUserTooLong, MaxUserLength)
	case runeCount(text) > MaxMessageLength:
		return Event{}, fmt.Errorf("%w: max %d characters", ErrMessageTooLong, MaxMessageLength)
	}

	if color == "" {
		color = DefaultUserColor
	}
	if !hexColorPattern.MatchString(color) {
		return Event{}, ErrInvalidColor
	}

	return Event{
		ID:        s.newID(),
		Sequence:  s.sequence.Add(1),
		Type:      EventTypeMessage,
		User:      user,
		Color:     strings.ToLower(color),
		Text:      text,
		CreatedAt: s.now().UTC(),
	}, nil
}

func (s *Service) SystemNotice(text string) Event {
	return Event{
		ID:        s.newID(),
		Sequence:  s.sequence.Add(1),
		Type:      EventTypeSystem,
		User:      systemUser,
		Color:     systemColor,
		Text:      strings.TrimSpace(text),
		CreatedAt: s.now().UTC(),
	}
}

func runeCount(value string) int {
	return len([]rune(value))
}
