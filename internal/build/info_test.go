package build

import "testing"

func TestNewInfoIncludesRuntimeMetadata(t *testing.T) {
	info := NewInfo("go-chat")

	if info.Service != "go-chat" {
		t.Fatalf("Service = %q, want go-chat", info.Service)
	}
	if info.Version == "" {
		t.Fatal("Version is empty")
	}
	if info.GoVersion == "" {
		t.Fatal("GoVersion is empty")
	}
}
