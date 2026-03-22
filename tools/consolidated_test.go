package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/feenlace/mcp-1c/dump"
	"github.com/feenlace/mcp-1c/onec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ---------- Tool definitions ----------

func TestCodeReadTool(t *testing.T) {
	tool := CodeReadTool()
	if tool.Name != "code_read" {
		t.Errorf("expected tool name %q, got %q", "code_read", tool.Name)
	}
	if tool.Description == "" {
		t.Error("expected non-empty description")
	}
	assertSchemaHasField(t, tool, "action")
	assertSchemaHasField(t, tool, "filter")
	assertSchemaHasField(t, tool, "object_type")
	assertSchemaHasField(t, tool, "object_name")
	assertSchemaHasField(t, tool, "form_name")
}

func TestCodeSearchTool(t *testing.T) {
	tool := CodeSearchTool()
	if tool.Name != "code_search" {
		t.Errorf("expected tool name %q, got %q", "code_search", tool.Name)
	}
	if tool.Description == "" {
		t.Error("expected non-empty description")
	}
	assertSchemaHasField(t, tool, "action")
	assertSchemaHasField(t, tool, "query")
	assertSchemaHasField(t, tool, "mode")
}

func TestCodeExecuteTool(t *testing.T) {
	tool := CodeExecuteTool()
	if tool.Name != "code_execute" {
		t.Errorf("expected tool name %q, got %q", "code_execute", tool.Name)
	}
	if tool.Description == "" {
		t.Error("expected non-empty description")
	}
	assertSchemaHasField(t, tool, "action")
	assertSchemaHasField(t, tool, "query")
	assertSchemaHasField(t, tool, "limit")
	assertSchemaHasField(t, tool, "parameters")
}

func TestSystemTool(t *testing.T) {
	tool := SystemTool()
	if tool.Name != "system" {
		t.Errorf("expected tool name %q, got %q", "system", tool.Name)
	}
	if tool.Description == "" {
		t.Error("expected non-empty description")
	}
	assertSchemaHasField(t, tool, "action")
	assertSchemaHasField(t, tool, "level")
	assertSchemaHasField(t, tool, "limit")
}

// ---------- code_read handler ----------

func TestCodeReadHandler_MetadataTree(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metadata" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"Справочники":["Контрагенты"],"Документы":["РеализацияТоваровУслуг"]}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer mockServer.Close()

	client := onec.NewClient(mockServer.URL, "", "")
	handler := NewCodeReadHandler(client, "")
	result, err := handler(context.Background(), makeReq(map[string]any{"action": "metadata_tree"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	if !strings.Contains(text, "Справочники") {
		t.Errorf("expected Справочники in result, got:\n%s", text)
	}
}

func TestCodeReadHandler_ConfigInfo(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/configuration" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"name":"Тест","version":"1.0","vendor":"Test","platform_version":"8.3","mode":"file"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer mockServer.Close()

	client := onec.NewClient(mockServer.URL, "", "")
	handler := NewCodeReadHandler(client, "")
	result, err := handler(context.Background(), makeReq(map[string]any{"action": "config_info"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	if !strings.Contains(text, "Тест") || !strings.Contains(text, "Файловый") {
		t.Errorf("expected config info in result, got:\n%s", text)
	}
}

func TestCodeReadHandler_ObjectStructure(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/object/Catalog/Контрагенты" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"name":"Контрагенты","synonym":"Контрагенты","attributes":[{"name":"ИНН","synonym":"ИНН","type":"Строка"}]}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer mockServer.Close()

	client := onec.NewClient(mockServer.URL, "", "")
	handler := NewCodeReadHandler(client, "")
	result, err := handler(context.Background(), makeReq(map[string]any{
		"action":      "object_structure",
		"object_type": "Catalog",
		"object_name": "Контрагенты",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	if !strings.Contains(text, "ИНН") {
		t.Errorf("expected ИНН in result, got:\n%s", text)
	}
}

func TestCodeReadHandler_UnknownAction(t *testing.T) {
	client := onec.NewClient("http://localhost:0", "", "")
	handler := NewCodeReadHandler(client, "")
	result, err := handler(context.Background(), makeReq(map[string]any{"action": "unknown"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for unknown action")
	}
	text := extractText(t, result)
	if !strings.Contains(text, "Неизвестное действие") {
		t.Errorf("expected error message, got:\n%s", text)
	}
}

// ---------- code_search handler ----------

func TestCodeSearchHandler_SyntaxHelp(t *testing.T) {
	handler := NewCodeSearchHandler(nil)
	result, err := handler(context.Background(), makeReq(map[string]any{
		"action": "syntax_help",
		"query":  "СтрНайти",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	if !strings.Contains(text, "СтрНайти") {
		t.Errorf("expected СтрНайти in result, got:\n%s", text)
	}
}

func TestCodeSearchHandler_SyntaxHelp_EmptyQuery(t *testing.T) {
	handler := NewCodeSearchHandler(nil)
	result, err := handler(context.Background(), makeReq(map[string]any{
		"action": "syntax_help",
		"query":  "",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for empty query")
	}
}

func TestCodeSearchHandler_TextNoDump(t *testing.T) {
	handler := NewCodeSearchHandler(nil)
	result, err := handler(context.Background(), makeReq(map[string]any{
		"action": "text",
		"query":  "ОбновитьЦены",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true when dump is not available")
	}
	text := extractText(t, result)
	if !strings.Contains(text, "недоступен") {
		t.Errorf("expected unavailability message, got:\n%s", text)
	}
}

func TestCodeSearchHandler_TextWithDump(t *testing.T) {
	dir := t.TempDir()
	mkBSL(t, dir, "Catalogs/Номенклатура/Ext/ObjectModule.bsl",
		"Процедура ОбновитьЦены()\n    // обновление\nКонецПроцедуры\n")

	index, err := dump.NewIndex(dir, false)
	if err != nil {
		t.Fatalf("NewIndex: %v", err)
	}
	defer index.Close()
	waitReady(t, index, 30*time.Second)

	handler := NewCodeSearchHandler(index)
	result, err := handler(context.Background(), makeReq(map[string]any{
		"action": "text",
		"query":  "ОбновитьЦены",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	if !strings.Contains(text, "ОбновитьЦены") {
		t.Errorf("expected search result, got:\n%s", text)
	}
}

func TestCodeSearchHandler_UnknownAction(t *testing.T) {
	handler := NewCodeSearchHandler(nil)
	result, err := handler(context.Background(), makeReq(map[string]any{
		"action": "unknown",
		"query":  "test",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for unknown action")
	}
}

// ---------- code_execute handler ----------

func TestCodeExecuteHandler_Query(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/query" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"columns":["Наименование"],"rows":[["Тест"]],"total":1,"truncated":false}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer mockServer.Close()

	client := onec.NewClient(mockServer.URL, "", "")
	handler := NewCodeExecuteHandler(client)
	result, err := handler(context.Background(), makeReq(map[string]any{
		"action": "query",
		"query":  "ВЫБРАТЬ Наименование ИЗ Справочник.Контрагенты",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	if !strings.Contains(text, "Тест") {
		t.Errorf("expected query result, got:\n%s", text)
	}
}

func TestCodeExecuteHandler_Validate(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/validate-query" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"valid":true,"errors":[]}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer mockServer.Close()

	client := onec.NewClient(mockServer.URL, "", "")
	handler := NewCodeExecuteHandler(client)
	result, err := handler(context.Background(), makeReq(map[string]any{
		"action": "validate",
		"query":  "ВЫБРАТЬ Наименование ИЗ Справочник.Контрагенты",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	if !strings.Contains(text, "корректен") {
		t.Errorf("expected valid result, got:\n%s", text)
	}
}

func TestCodeExecuteHandler_UnknownAction(t *testing.T) {
	client := onec.NewClient("http://localhost:0", "", "")
	handler := NewCodeExecuteHandler(client)
	result, err := handler(context.Background(), makeReq(map[string]any{
		"action": "unknown",
		"query":  "test",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for unknown action")
	}
}

// ---------- system handler ----------

func TestSystemHandler_EventLog(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/eventlog" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"events":[{"date":"2026-03-07T10:00:00","level":"Ошибка","event":"Тест","user":"Админ"}],"total":1}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer mockServer.Close()

	client := onec.NewClient(mockServer.URL, "", "")
	handler := NewSystemHandler(client)
	result, err := handler(context.Background(), makeReq(map[string]any{
		"action": "event_log",
		"level":  "Ошибка",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	if !strings.Contains(text, "Журнал регистрации") {
		t.Errorf("expected event log in result, got:\n%s", text)
	}
}

func TestSystemHandler_UnknownAction(t *testing.T) {
	client := onec.NewClient("http://localhost:0", "", "")
	handler := NewSystemHandler(client)
	result, err := handler(context.Background(), makeReq(map[string]any{"action": "unknown"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for unknown action")
	}
}

// ---------- errResult ----------

func TestErrResult(t *testing.T) {
	result := errResult("тестовая ошибка")
	if !result.IsError {
		t.Error("expected IsError=true")
	}
	text := extractText(t, result)
	if !strings.Contains(text, "тестовая ошибка") {
		t.Errorf("expected error message, got:\n%s", text)
	}
}

// ---------- helpers ----------

func makeReq(args map[string]any) *mcp.CallToolRequest {
	raw, _ := json.Marshal(args)
	return &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Arguments: raw,
		},
	}
}

func extractText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
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
	return tc.Text
}

func assertSchemaHasField(t *testing.T, tool *mcp.Tool, field string) {
	t.Helper()
	raw, ok := tool.InputSchema.(json.RawMessage)
	if !ok {
		t.Fatalf("expected json.RawMessage InputSchema, got %T", tool.InputSchema)
	}
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("parsing input schema: %v", err)
	}
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("expected properties in schema for field %q", field)
	}
	if _, ok := props[field]; !ok {
		t.Errorf("missing property %q in tool %q schema", field, tool.Name)
	}
}
