package websocket

import "time"

type rateLimiter struct {
	limit       int
	window      time.Duration
	now         func() time.Time
	windowStart time.Time
	count       int
}

func newRateLimiter(limit int, window time.Duration, now func() time.Time) *rateLimiter {
	if now == nil {
		now = time.Now
	}
	return &rateLimiter{
		limit:  limit,
		window: window,
		now:    now,
	}
}

func (l *rateLimiter) Allow() bool {
	if l == nil || l.limit <= 0 || l.window <= 0 {
		return true
	}

	now := l.now()
	if l.windowStart.IsZero() || now.Sub(l.windowStart) >= l.window {
		l.windowStart = now
		l.count = 0
	}

	if l.count >= l.limit {
		return false
	}

	l.count++
	return true
}
