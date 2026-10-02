package discord

import (
	"strings"
	"testing"
)

func TestClientBlacklist(t *testing.T) {
	client := &Client{
		DefaultGuildID: "1001",
		Blacklist: map[string]struct{}{
			"9999": {},
			"8888": {},
		},
	}

	if !client.IsBlacklisted("9999") {
		t.Fatal("expected guild 9999 to be blacklisted")
	}
	if client.IsBlacklisted("1001") {
		t.Fatal("guild 1001 should not be blacklisted")
	}

	err := client.CheckGuildAllowed("9999")
	if err == nil || !strings.Contains(err.Error(), "blacklisted") {
		t.Fatalf("expected blacklist error, got %v", err)
	}

	errAllowed := client.CheckGuildAllowed("1001")
	if errAllowed != nil {
		t.Fatalf("expected allowed guild, got %v", errAllowed)
	}
}

func TestResolveGuildID(t *testing.T) {
	client := &Client{
		DefaultGuildID: "1001",
		Blacklist: map[string]struct{}{
			"9999": {},
		},
	}

	resolved, err := client.ResolveGuildID("")
	if err != nil {
		t.Fatalf("unexpected error resolving default guild: %v", err)
	}
	if resolved != "1001" {
		t.Fatalf("expected 1001, got %s", resolved)
	}

	overrideResolved, err := client.ResolveGuildID("2002")
	if err != nil {
		t.Fatalf("unexpected error resolving override guild: %v", err)
	}
	if overrideResolved != "2002" {
		t.Fatalf("expected 2002, got %s", overrideResolved)
	}

	_, errBlacklisted := client.ResolveGuildID("9999")
	if errBlacklisted == nil {
		t.Fatal("expected error resolving blacklisted guild override")
	}

	clientEmpty := &Client{
		DefaultGuildID: "",
		Blacklist:      map[string]struct{}{},
	}
	_, errMissing := clientEmpty.ResolveGuildID("")
	if errMissing == nil {
		t.Fatal("expected error when no default guild and no override passed")
	}
}
