package mcp

import (
	"fmt"
	"log"

	"github.com/0xOpCode/discord-mcp-go/internal/discord"
	"github.com/mark3labs/mcp-go/server"
)

type Server struct {
	MCPServer *server.MCPServer
	Client    *discord.Client
}

func NewServer(client *discord.Client) *Server {
	s := server.NewMCPServer(
		"discord-mcp-go",
		"1.0.0",
		server.WithLogging(),
	)

	RegisterServerTools(s, client)
	RegisterMessageTools(s, client)
	RegisterUserTools(s, client)

	return &Server{
		MCPServer: s,
		Client:    client,
	}
}

func (s *Server) Serve(transport, port string) error {
	switch transport {
	case "stdio":
		log.Println("Starting discord-mcp-go in STDIO mode")
		return server.ServeStdio(s.MCPServer)
	case "sse":
		fallthrough
	default:
		baseURL := fmt.Sprintf("http://localhost:%s", port)
		log.Printf("Starting discord-mcp-go in HTTP SSE mode on :%s (baseURL: %s)", port, baseURL)
		sseServer := server.NewSSEServer(s.MCPServer, baseURL)
		return sseServer.Start(":" + port)
	}
}
