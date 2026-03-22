package server

import (
	"github.com/feenlace/mcp-1c/dump"
	"github.com/feenlace/mcp-1c/onec"
	"github.com/feenlace/mcp-1c/prompts"
	"github.com/feenlace/mcp-1c/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// New creates an MCP server with basic configuration and registers tools.
// If dumpIndex is provided, the code_search text action will be available.
func New(version string, onecClient *onec.Client, dumpIndex *dump.Index) *mcp.Server {
	s := mcp.NewServer(
		&mcp.Implementation{
			Name:    "mcp-1c",
			Version: version,
		},
		nil,
	)

	// Pass dump directory to code_read handler so it can enrich form responses
	// with data from Form.xml files parsed from the dump.
	var dumpDir string
	if dumpIndex != nil {
		dumpDir = dumpIndex.Dir()
	}

	s.AddTool(tools.CodeReadTool(), tools.NewCodeReadHandler(onecClient, dumpDir))
	s.AddTool(tools.CodeSearchTool(), tools.NewCodeSearchHandler(dumpIndex))
	s.AddTool(tools.CodeExecuteTool(), tools.NewCodeExecuteHandler(onecClient))
	s.AddTool(tools.SystemTool(), tools.NewSystemHandler(onecClient))

	prompts.RegisterAll(s)
	return s
}
