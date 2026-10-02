package mcp

import (
	"fmt"
	"log"
	"net/http"

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
		"1.1.0",
		server.WithLogging(),
	)

	RegisterServerTools(s, client)
	RegisterMessageTools(s, client)
	RegisterUserTools(s, client)
	RegisterChannelTools(s, client)
	RegisterCategoryTools(s, client)
	RegisterRoleTools(s, client)
	RegisterPermissionTools(s, client)
	RegisterModerationTools(s, client)
	RegisterVoiceTools(s, client)
	RegisterWebhookTools(s, client)
	RegisterEventTools(s, client)
	RegisterInviteTools(s, client)
	RegisterForumTools(s, client)
	RegisterEmojiTools(s, client)
	RegisterAutoModTools(s, client)
	RegisterPollTools(s, client)
	RegisterSoundboardTools(s, client)
	RegisterStickerTools(s, client)
	RegisterRoleConnectionTools(s, client)
	RegisterPipelineTools(s, client)

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
		log.Printf("Starting discord-mcp-go in HTTP mode on :%s (baseURL: %s)", port, baseURL)
		httpServer := NewHTTPServer(s, baseURL)
		srv := &http.Server{
			Addr:    ":" + port,
			Handler: httpServer.Handler(),
		}
		return srv.ListenAndServe()
	}
}
