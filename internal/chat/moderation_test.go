package chat

import (
	"errors"
	"testing"
)

func TestBannedTermsModeratorBlocksTerms(t *testing.T) {
	moderator := NewBannedTermsModerator([]string{"spam"})

	err := moderator.Validate(MessageInput{Text: "this is SPAM"})
	if !errors.Is(err, ErrMessageBlocked) {
		t.Fatalf("Validate() error = %v, want %v", err, ErrMessageBlocked)
	}
}

func TestBannedTermsModeratorAllowsCleanText(t *testing.T) {
	moderator := NewBannedTermsModerator([]string{"spam"})

	if err := moderator.Validate(MessageInput{Text: "hello"}); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}
