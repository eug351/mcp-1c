package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/feenlace/mcp-1c/internal/onec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type moduleCodeInput struct {
	ObjectType string `json:"object_type"`
	ObjectName string `json:"object_name"`
	ModuleKind string `json:"module_kind"`
}

// ModuleCodeTool returns the MCP tool definition for get_module_code.
func ModuleCodeTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "get_module_code",
		Description: "Получить исходный код модуля объекта 1С. Используй когда нужно увидеть логику обработки проведения документа, событий справочника, код общего модуля и т.д.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"object_type": {
					"type": "string",
					"description": "Тип объекта: Document, Catalog, CommonModule, InformationRegister, AccumulationRegister, AccountingRegister"
				},
				"object_name": {
					"type": "string",
					"description": "Имя объекта, например РеализацияТоваровУслуг"
				},
				"module_kind": {
					"type": "string",
					"description": "Вид модуля: ObjectModule, ManagerModule, FormModule, CommonModule"
				}
			},
			"required": ["object_type", "object_name", "module_kind"]
		}`),
	}
}

// NewModuleCodeHandler returns a ToolHandler that fetches module source code from 1C.
func NewModuleCodeHandler(client *onec.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input moduleCodeInput
		if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
			return nil, fmt.Errorf("parsing input: %w", err)
		}
		if input.ObjectType == "" || input.ObjectName == "" || input.ModuleKind == "" {
			return nil, fmt.Errorf("object_type, object_name and module_kind are required")
		}

		endpoint := fmt.Sprintf("/module/%s/%s/%s", input.ObjectType, input.ObjectName, input.ModuleKind)
		var mod onec.ModuleCode
		if err := client.Get(ctx, endpoint, &mod); err != nil {
			return nil, fmt.Errorf("fetching module code from 1C: %w", err)
		}

		return textResult(formatModuleCode(&mod)), nil
	}
}

func formatModuleCode(mod *onec.ModuleCode) string {
	var b strings.Builder

	fmt.Fprintf(&b, "## Модуль %s: %s\n\n", mod.ModuleKind, mod.Name)

	if mod.Code == "" {
		b.WriteString("Модуль пуст.\n")
		return b.String()
	}

	b.WriteString("```bsl\n")
	b.WriteString(mod.Code)
	if !strings.HasSuffix(mod.Code, "\n") {
		b.WriteByte('\n')
	}
	b.WriteString("```\n")

	return b.String()
}
