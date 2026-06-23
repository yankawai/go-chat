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

func TestHistoryListAfterFiltersBySequence(t *testing.T) {
	history := NewHistory(10)
	history.Append(Event{ID: "1", Sequence: 1})
	history.Append(Event{ID: "2", Sequence: 2})
	history.Append(Event{ID: "3", Sequence: 3})

	events := history.ListAfter(1, 0)
	if len(events) != 2 {
		t.Fatalf("len(events) = %d, want 2", len(events))
	}
	if events[0].ID != "2" || events[1].ID != "3" {
		t.Fatalf("events = %+v, want ids 2 and 3", events)
	}
}

func TestHistoryListAfterAppliesLimit(t *testing.T) {
	history := NewHistory(10)
	history.Append(Event{ID: "1", Sequence: 1})
	history.Append(Event{ID: "2", Sequence: 2})
	history.Append(Event{ID: "3", Sequence: 3})

	events := history.ListAfter(0, 1)
	if len(events) != 1 || events[0].ID != "3" {
		t.Fatalf("events = %+v, want only id 3", events)
	}
}
