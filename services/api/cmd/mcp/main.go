// Command mcp serves the mytax draft relief catalog over MCP on stdin/stdout.
//
// Local, read-only developer tool: stdio transport only, draft data only, no evaluation,
// no network listener and no persistence. It is not for deployment.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fikrimastor/mytax/services/api/internal/mcpserver"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// log defaults to stderr; stdout belongs to the MCP wire.
	if err := mcpserver.New("0.1.0-dev").Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
