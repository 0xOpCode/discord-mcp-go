package config

import (
	"os"
	"strings"
)

type Config struct {
	DiscordToken      string
	DefaultGuildID    string
	Port              string
	Transport         string
	BlacklistedGuilds []string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	transport := strings.ToLower(os.Getenv("TRANSPORT"))
	if transport != "stdio" {
		transport = "sse"
	}

	var blacklisted []string
	if raw := os.Getenv("DISCORD_BLACKLISTED_GUILDS"); raw != "" {
		for _, id := range strings.Split(raw, ",") {
			trimmed := strings.TrimSpace(id)
			if trimmed != "" {
				blacklisted = append(blacklisted, trimmed)
			}
		}
	}

	return &Config{
		DiscordToken:      os.Getenv("DISCORD_TOKEN"),
		DefaultGuildID:    os.Getenv("DISCORD_GUILD_ID"),
		Port:              port,
		Transport:         transport,
		BlacklistedGuilds: blacklisted,
	}
}
