package config

import (
	"os"
	"reflect"
	"testing"
)

func TestConfigLoadDefaults(t *testing.T) {
	os.Clearenv()

	cfg := Load()

	if cfg.Port != "8085" {
		t.Fatalf("expected default port 8085, got %s", cfg.Port)
	}
	if cfg.Transport != "sse" {
		t.Fatalf("expected default transport sse, got %s", cfg.Transport)
	}
	if len(cfg.BlacklistedGuilds) != 0 {
		t.Fatalf("expected empty blacklisted guilds, got %v", cfg.BlacklistedGuilds)
	}
	if len(cfg.AllowedFilePaths) != 0 {
		t.Fatalf("expected empty allowed file paths, got %v", cfg.AllowedFilePaths)
	}
}

func TestConfigLoadCustomEnv(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("TRANSPORT", "stdio")
	os.Setenv("DISCORD_TOKEN", "test-token-xyz")
	os.Setenv("DISCORD_GUILD_ID", "1122334455")
	os.Setenv("DISCORD_BLACKLISTED_GUILDS", "111, 222 , 333")
	os.Setenv("DISCORD_ALLOWED_FILE_PATHS", "/tmp, /var/log")
	defer os.Clearenv()

	cfg := Load()

	if cfg.Port != "9090" {
		t.Fatalf("expected port 9090, got %s", cfg.Port)
	}
	if cfg.Transport != "stdio" {
		t.Fatalf("expected transport stdio, got %s", cfg.Transport)
	}
	if cfg.DiscordToken != "test-token-xyz" {
		t.Fatalf("expected token test-token-xyz, got %s", cfg.DiscordToken)
	}
	if cfg.DefaultGuildID != "1122334455" {
		t.Fatalf("expected guild id 1122334455, got %s", cfg.DefaultGuildID)
	}

	expectedBlacklist := []string{"111", "222", "333"}
	if !reflect.DeepEqual(cfg.BlacklistedGuilds, expectedBlacklist) {
		t.Fatalf("expected blacklist %v, got %v", expectedBlacklist, cfg.BlacklistedGuilds)
	}

	if len(cfg.AllowedFilePaths) != 2 || cfg.AllowedFilePaths[0] != "/tmp" || cfg.AllowedFilePaths[1] != "/var/log" {
		t.Fatalf("expected allowed file paths [/tmp /var/log], got %v", cfg.AllowedFilePaths)
	}
}
