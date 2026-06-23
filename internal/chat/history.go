package chat

import "sync"

const DefaultHistoryLimit = 100

type History struct {
	mu     sync.RWMutex
	limit  int
	events []Event
}

func NewHistory(limit int) *History {
	if limit <= 0 {
		limit = DefaultHistoryLimit
	}

	return &History{
		limit:  limit,
		events: make([]Event, 0, limit),
	}
}

func (h *History) Append(event Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.events = append(h.events, event)
	if len(h.events) > h.limit {
		h.events = h.events[len(h.events)-h.limit:]
	}
}

func (h *History) List(limit int) []Event {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if limit <= 0 || limit > len(h.events) {
		limit = len(h.events)
	}

	start := len(h.events) - limit
	events := make([]Event, limit)
	copy(events, h.events[start:])
	return events
}

func (h *History) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.events)
}
