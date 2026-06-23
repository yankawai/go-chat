package chat

import "errors"

var ErrMessageBlocked = errors.New("message was blocked by moderation")

type Moderator interface {
	Validate(MessageInput) error
}

type ModeratorFunc func(MessageInput) error

func (f ModeratorFunc) Validate(input MessageInput) error {
	return f(input)
}
