package websocket

import "testing"

func TestOriginAllowed(t *testing.T) {
	if !originAllowed("https://chat.example.com", []string{"https://chat.example.com"}) {
		t.Fatal("originAllowed() = false, want true")
	}
	if originAllowed("https://evil.example.com", []string{"https://chat.example.com"}) {
		t.Fatal("originAllowed() = true, want false")
	}
}

func TestSameHostOrigin(t *testing.T) {
	if !sameHostOrigin("http://localhost:8080", "localhost:8080") {
		t.Fatal("sameHostOrigin() = false, want true")
	}
	if sameHostOrigin("http://localhost:3000", "localhost:8080") {
		t.Fatal("sameHostOrigin() = true, want false")
	}
}

func TestIsLoopbackOrigin(t *testing.T) {
	tests := []struct {
		origin string
		want   bool
	}{
		{origin: "http://localhost:3000", want: true},
		{origin: "http://127.0.0.1:3000", want: true},
		{origin: "http://[::1]:3000", want: true},
		{origin: "https://chat.example.com", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.origin, func(t *testing.T) {
			if got := isLoopbackOrigin(tt.origin); got != tt.want {
				t.Fatalf("isLoopbackOrigin() = %v, want %v", got, tt.want)
			}
		})
	}
}
