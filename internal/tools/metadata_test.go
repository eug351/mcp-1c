package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/feenlace/mcp-1c/internal/onec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMetadataHandler(t *testing.T) {
	// Start a mock 1C server.
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/metadata" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"Справочники":["Контрагенты","Номенклатура"],"Документы":["РеализацияТоваровУслуг"],"Регистры":["КурсыВалют"]}`))
	}))
	defer mockServer.Close()

	client := onec.NewClient(mockServer.URL)
	handler := NewMetadataHandler(client)

	result, err := handler(context.Background(), &mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Content) == 0 {
		t.Fatal("expected non-empty content")
	}

	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if tc.Text == "" {
		t.Fatal("expected non-empty text")
	}

	// Verify the text contains key metadata items.
	for _, want := range []string{"Контрагенты", "Номенклатура", "РеализацияТоваровУслуг", "КурсыВалют"} {
		if !contains(tc.Text, want) {
			t.Errorf("expected text to contain %q, got:\n%s", want, tc.Text)
		}
	}
}

func TestMetadataTool(t *testing.T) {
	tool := MetadataTool()
	if tool == nil {
		t.Fatal("expected non-nil tool")
	}
	if tool.Name != "get_metadata_tree" {
		t.Errorf("expected tool name %q, got %q", "get_metadata_tree", tool.Name)
	}
	if tool.Description == "" {
		t.Error("expected non-empty description")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
