package tools

import "github.com/modelcontextprotocol/go-sdk/mcp"

// textResult wraps a text string into an MCP tool result.
func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}
}
