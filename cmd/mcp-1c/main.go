package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/feenlace/mcp-1c/internal/onec"
	"github.com/feenlace/mcp-1c/internal/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	baseURL := flag.String("base", "http://localhost:8080/mcp", "Base URL of 1C HTTP service")
	flag.Parse()

	client := onec.NewClient(*baseURL)
	s := server.New(client)

	if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintf(os.Stderr, "mcp-1c error: %v\n", err)
		os.Exit(1)
	}
}
