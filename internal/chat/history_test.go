package chat

import "testing"

func TestHistoryKeepsBoundedEvents(t *testing.T) {
	history := NewHistory(2)
	history.Append(Event{ID: "1"})
	history.Append(Event{ID: "2"})
	history.Append(Event{ID: "3"})

	events := history.List(0)
	if len(events) != 2 {
		t.Fatalf("len(events) = %d, want 2", len(events))
	}
	if events[0].ID != "2" || events[1].ID != "3" {
		t.Fatalf("events = %+v, want ids 2 and 3", events)
	}
}

func TestHistoryListReturnsCopy(t *testing.T) {
	history := NewHistory(2)
	history.Append(Event{ID: "1"})

	events := history.List(1)
	events[0].ID = "mutated"

	events = history.List(1)
	if events[0].ID != "1" {
		t.Fatalf("history was mutated through returned slice")
	}
}
