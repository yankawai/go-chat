package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatusRecorderTracksStatusAndBytes(t *testing.T) {
	rec := httptest.NewRecorder()
	status := newStatusRecorder(rec)

	status.WriteHeader(http.StatusCreated)
	n, err := status.Write([]byte("created"))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if n != len("created") {
		t.Fatalf("bytes written = %d, want %d", n, len("created"))
	}
	if status.status() != http.StatusCreated {
		t.Fatalf("status = %d, want %d", status.status(), http.StatusCreated)
	}
	if status.bytesWritten != len("created") {
		t.Fatalf("bytesWritten = %d, want %d", status.bytesWritten, len("created"))
	}
}

func TestStatusRecorderDefaultsToOK(t *testing.T) {
	status := newStatusRecorder(httptest.NewRecorder())

	if status.status() != http.StatusOK {
		t.Fatalf("status = %d, want %d", status.status(), http.StatusOK)
	}
}
