package websocket

import (
	"testing"
	"time"
)

func TestRateLimiterBlocksAfterLimit(t *testing.T) {
	now := time.Date(2026, 6, 23, 10, 0, 0, 0, time.UTC)
	limiter := newRateLimiter(2, time.Second, func() time.Time { return now })

	if !limiter.Allow() || !limiter.Allow() {
		t.Fatal("first two events should be allowed")
	}
	if limiter.Allow() {
		t.Fatal("third event should be blocked")
	}
}

func TestRateLimiterResetsAfterWindow(t *testing.T) {
	now := time.Date(2026, 6, 23, 10, 0, 0, 0, time.UTC)
	limiter := newRateLimiter(1, time.Second, func() time.Time { return now })

	if !limiter.Allow() {
		t.Fatal("first event should be allowed")
	}
	now = now.Add(time.Second)
	if !limiter.Allow() {
		t.Fatal("event after reset should be allowed")
	}
}
