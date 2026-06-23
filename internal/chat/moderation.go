package chat

import (
	"errors"
	"strings"
)

var ErrMessageBlocked = errors.New("message was blocked by moderation")

type Moderator interface {
	Validate(MessageInput) error
}

type ModeratorFunc func(MessageInput) error

func (f ModeratorFunc) Validate(input MessageInput) error {
	return f(input)
}

type BannedTermsModerator struct {
	terms []string
}

func NewBannedTermsModerator(terms []string) *BannedTermsModerator {
	normalized := make([]string, 0, len(terms))
	for _, term := range terms {
		term = strings.ToLower(strings.TrimSpace(term))
		if term != "" {
			normalized = append(normalized, term)
		}
	}

	return &BannedTermsModerator{terms: normalized}
}

func (m *BannedTermsModerator) Validate(input MessageInput) error {
	if m == nil || len(m.terms) == 0 {
		return nil
	}

	text := strings.ToLower(input.Text)
	for _, term := range m.terms {
		if strings.Contains(text, term) {
			return ErrMessageBlocked
		}
	}

	return nil
}
