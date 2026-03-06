package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/feenlace/mcp-1c/internal/onec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestModuleCodeHandler(t *testing.T) {
	const mockResponse = `{
		"Имя": "РеализацияТоваровУслуг",
		"ВидМодуля": "ObjectModule",
		"Код": "Процедура ОбработкаПроведения(Отказ, РежимПроведения)\n\t// Движения по регистрам\nКонецПроцедуры"
	}`

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/module/Document/РеализацияТоваровУслуг/ObjectModule" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer mockServer.Close()

	client := onec.NewClient(mockServer.URL, "", "")
	handler := NewModuleCodeHandler(client)

	args, _ := json.Marshal(map[string]string{
		"object_type": "Document",
		"object_name": "РеализацияТоваровУслуг",
		"module_kind": "ObjectModule",
	})
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "get_module_code",
			Arguments: args,
		},
	}

	result, err := handler(context.Background(), req)
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

	for _, want := range []string{
		"РеализацияТоваровУслуг",
		"ObjectModule",
		"ОбработкаПроведения",
		"Движения по регистрам",
		"КонецПроцедуры",
		"```bsl",
	} {
		if !strings.Contains(tc.Text, want) {
			t.Errorf("expected text to contain %q, got:\n%s", want, tc.Text)
		}
	}
}

func TestModuleCodeHandler_EmptyModule(t *testing.T) {
	const mockResponse = `{
		"Имя": "СправочникТест",
		"ВидМодуля": "ManagerModule",
		"Код": ""
	}`

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer mockServer.Close()

	client := onec.NewClient(mockServer.URL, "", "")
	handler := NewModuleCodeHandler(client)

	args, _ := json.Marshal(map[string]string{
		"object_type": "Catalog",
		"object_name": "СправочникТест",
		"module_kind": "ManagerModule",
	})
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "get_module_code",
			Arguments: args,
		},
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(tc.Text, "пуст") {
		t.Errorf("expected text to contain 'пуст', got:\n%s", tc.Text)
	}
}

func TestModuleCodeTool(t *testing.T) {
	tool := ModuleCodeTool()
	if tool == nil {
		t.Fatal("expected non-nil tool")
	}
	if tool.Name != "get_module_code" {
		t.Errorf("expected tool name %q, got %q", "get_module_code", tool.Name)
	}
	if tool.Description == "" {
		t.Error("expected non-empty description")
	}
}
