package prompts

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestRegisterAll(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, nil)
	RegisterAll(s)

	ctx := context.Background()
	ct, st := mcp.NewInMemoryTransports()

	_, err := s.Connect(ctx, st, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	session, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer session.Close()

	result, err := session.ListPrompts(ctx, nil)
	if err != nil {
		t.Fatalf("ListPrompts error: %v", err)
	}

	if len(result.Prompts) != 8 {
		names := make([]string, len(result.Prompts))
		for i, p := range result.Prompts {
			names[i] = p.Name
		}
		t.Fatalf("expected 8 prompts, got %d: %v", len(result.Prompts), names)
	}

	expected := map[string]bool{
		"review_module":   false,
		"write_posting":   false,
		"optimize_query":  false,
		"explain_config":  false,
		"analyze_error":   false,
		"find_duplicates": false,
		"write_report":    false,
		"explain_object":  false,
	}
	for _, p := range result.Prompts {
		if _, ok := expected[p.Name]; !ok {
			t.Errorf("unexpected prompt name: %s", p.Name)
		}
		expected[p.Name] = true
	}
	for name, found := range expected {
		if !found {
			t.Errorf("prompt %q not found in ListPrompts result", name)
		}
	}
}

func TestPromptHandlers(t *testing.T) {
	tests := []struct {
		name        string
		handler     mcp.PromptHandler
		arguments   map[string]string
		wantKeyword string
	}{
		{
			name:    "review_module",
			handler: handleReviewModule,
			arguments: map[string]string{
				"object_type": "Document",
				"object_name": "ПриходнаяНакладная",
			},
			wantKeyword: "get_object_structure",
		},
		{
			name:    "write_posting",
			handler: handleWritePosting,
			arguments: map[string]string{
				"document_name": "РеализацияТоваровУслуг",
			},
			wantKeyword: "ОбработкаПроведения",
		},
		{
			name:    "optimize_query",
			handler: handleOptimizeQuery,
			arguments: map[string]string{
				"query": "ВЫБРАТЬ * ИЗ Справочник.Контрагенты",
			},
			wantKeyword: "execute_query",
		},
		{
			name:        "explain_config",
			handler:     handleExplainConfig,
			arguments:   map[string]string{},
			wantKeyword: "get_metadata_tree",
		},
		{
			name:    "analyze_error",
			handler: handleAnalyzeError,
			arguments: map[string]string{
				"error_text": "Поле не найдено \"Номенклатура\"",
			},
			wantKeyword: "get_module_code",
		},
		{
			name:    "find_duplicates",
			handler: handleFindDuplicates,
			arguments: map[string]string{
				"object_type": "Catalog",
				"object_name": "Контрагенты",
			},
			wantKeyword: "get_module_code",
		},
		{
			name:    "write_report",
			handler: handleWriteReport,
			arguments: map[string]string{
				"description": "Отчёт по продажам за период",
			},
			wantKeyword: "execute_query",
		},
		{
			name:    "explain_object",
			handler: handleExplainObject,
			arguments: map[string]string{
				"object_type": "AccumulationRegister",
				"object_name": "ТоварыНаСкладах",
			},
			wantKeyword: "get_object_structure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &mcp.GetPromptRequest{
				Params: &mcp.GetPromptParams{
					Name:      tt.name,
					Arguments: tt.arguments,
				},
			}

			result, err := tt.handler(context.Background(), req)
			if err != nil {
				t.Fatalf("handler returned error: %v", err)
			}

			if result.Description == "" {
				t.Error("expected non-empty Description")
			}

			if len(result.Messages) != 1 {
				t.Fatalf("expected 1 message, got %d", len(result.Messages))
			}

			msg := result.Messages[0]
			if msg.Role != "user" {
				t.Errorf("expected role \"user\", got %q", msg.Role)
			}

			tc, ok := msg.Content.(*mcp.TextContent)
			if !ok {
				t.Fatalf("expected *mcp.TextContent, got %T", msg.Content)
			}

			if tc.Text == "" {
				t.Error("expected non-empty text content")
			}

			if len(tc.Text) < 50 {
				t.Errorf("text content too short (%d chars), expected detailed instructions", len(tc.Text))
			}

			if !containsSubstring(tc.Text, tt.wantKeyword) {
				t.Errorf("text content does not contain expected keyword %q:\n%s", tt.wantKeyword, tc.Text)
			}
		})
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && contains(s, substr))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
