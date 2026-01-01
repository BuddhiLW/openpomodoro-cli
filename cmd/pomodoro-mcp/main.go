package main

import (
	"log"

	pomodoromcp "github.com/BuddhiLW/openpomodoro-cli/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create the MCP server
	s := pomodoromcp.NewServer()

	// Start the server with stdio transport
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
