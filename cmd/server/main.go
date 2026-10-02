package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/0xOpCode/discord-mcp-go/internal/config"
	"github.com/0xOpCode/discord-mcp-go/internal/discord"
	"github.com/0xOpCode/discord-mcp-go/internal/mcp"
)

func main() {
	cfg := config.Load()

	transportFlag := flag.String("transport", cfg.Transport, "MCP transport protocol: 'sse' or 'stdio'")
	portFlag := flag.String("port", cfg.Port, "HTTP port for SSE transport")
	tokenFlag := flag.String("token", cfg.DiscordToken, "Discord Bot Token")
	guildIDFlag := flag.String("guild-id", cfg.DefaultGuildID, "Default Discord Server (Guild) ID")
	blacklistedFlag := flag.String("blacklisted-guilds", "", "Comma-separated list of blacklisted Guild IDs")
	flag.Parse()

	token := *tokenFlag
	if token == "" {
		token = os.Getenv("DISCORD_TOKEN")
	}
	if token == "" {
		log.Fatal("Discord token must be provided via -token flag or DISCORD_TOKEN environment variable")
	}

	guildID := *guildIDFlag
	if guildID == "" {
		guildID = os.Getenv("DISCORD_GUILD_ID")
	}

	blacklisted := cfg.BlacklistedGuilds
	if *blacklistedFlag != "" {
		for _, id := range strings.Split(*blacklistedFlag, ",") {
			trimmed := strings.TrimSpace(id)
			if trimmed != "" {
				blacklisted = append(blacklisted, trimmed)
			}
		}
	}

	discordClient, err := discord.NewClient(token, guildID, blacklisted)
	if err != nil {
		log.Fatalf("Discord initialization failed: %v", err)
	}

	log.Printf("Authenticated as Discord bot: %s (ID: %s)", discordClient.BotUser.Username, discordClient.BotUser.ID)

	mcpServer := mcp.NewServer(discordClient)

	if err := mcpServer.Serve(*transportFlag, *portFlag); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
