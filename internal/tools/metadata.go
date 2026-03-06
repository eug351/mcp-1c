package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/feenlace/mcp-1c/internal/onec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MetadataTool returns the MCP tool definition for get_metadata_tree.
func MetadataTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "get_metadata_tree",
		Description: "Получить дерево метаданных конфигурации 1С: список справочников, документов и регистров. Используй когда нужно узнать структуру конфигурации, какие объекты есть в базе.",
		InputSchema: json.RawMessage(`{"type":"object"}`),
	}
}

// NewMetadataHandler returns a ToolHandler that fetches the metadata tree from 1C.
func NewMetadataHandler(client *onec.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var tree onec.MetadataTree
		if err := client.Get(ctx, "/metadata", &tree); err != nil {
			return nil, fmt.Errorf("fetching metadata from 1C: %w", err)
		}

		text := formatMetadataTree(&tree)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: text},
			},
		}, nil
	}
}

// formatMetadataTree formats the metadata tree as markdown text.
func formatMetadataTree(tree *onec.MetadataTree) string {
	var b strings.Builder
	b.WriteString("# Метаданные конфигурации 1С\n\n")

	sections := []struct {
		title string
		items []string
	}{
		{"Справочники", tree.Catalogs},
		{"Документы", tree.Documents},
		{"Регистры сведений", tree.InformationRegisters},
		{"Регистры накопления", tree.AccumulationRegisters},
		{"Регистры бухгалтерии", tree.AccountingRegisters},
		{"Общие модули", tree.CommonModules},
	}

	for _, s := range sections {
		if len(s.items) == 0 {
			continue
		}
		fmt.Fprintf(&b, "## %s\n", s.title)
		for _, name := range s.items {
			b.WriteString("- ")
			b.WriteString(name)
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}

	return b.String()
}
