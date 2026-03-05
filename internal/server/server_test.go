package server

import (
	"testing"

	"github.com/feenlace/mcp-1c/internal/onec"
)

func TestNewServer(t *testing.T) {
	client := onec.NewClient("http://localhost:8080/mcp")
	s := New(client)
	if s == nil {
		t.Fatal("expected non-nil server")
	}
}
