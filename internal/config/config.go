package config

import (
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	DiscordToken      string
	DefaultGuildID    string
	Port              string
	Transport         string
	BlacklistedGuilds []string
	AllowedFilePaths  []string
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

	var allowedPaths []string
	if raw := os.Getenv("DISCORD_ALLOWED_FILE_PATHS"); raw != "" {
		for _, p := range strings.Split(raw, ",") {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				if abs, err := filepath.Abs(trimmed); err == nil {
					allowedPaths = append(allowedPaths, filepath.Clean(abs))
				} else {
					allowedPaths = append(allowedPaths, filepath.Clean(trimmed))
				}
			}
		}
	}

	return &Config{
		DiscordToken:      os.Getenv("DISCORD_TOKEN"),
		DefaultGuildID:    os.Getenv("DISCORD_GUILD_ID"),
		Port:              port,
		Transport:         transport,
		BlacklistedGuilds: blacklisted,
		AllowedFilePaths:  allowedPaths,
	}
}
