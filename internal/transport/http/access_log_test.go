package http

import (
	"bufio"
	"net"
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

func TestStatusRecorderRequiresHijackerSupport(t *testing.T) {
	status := newStatusRecorder(httptest.NewRecorder())

	_, _, err := status.Hijack()
	if err == nil {
		t.Fatal("Hijack() error = nil, want unsupported error")
	}
}

func TestStatusRecorderPassesThroughHijacker(t *testing.T) {
	status := newStatusRecorder(fakeHijackerResponseWriter{})

	_, _, err := status.Hijack()
	if err != nil {
		t.Fatalf("Hijack() error = %v, want nil", err)
	}
}

type fakeHijackerResponseWriter struct{}

func (fakeHijackerResponseWriter) Header() http.Header {
	return http.Header{}
}

func (fakeHijackerResponseWriter) Write(body []byte) (int, error) {
	return len(body), nil
}

func (fakeHijackerResponseWriter) WriteHeader(int) {}

func (fakeHijackerResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, nil
}
