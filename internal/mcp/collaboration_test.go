package mcp

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/0xOpCode/discord-mcp-go/internal/discord"
	"github.com/0xOpCode/discord-mcp-go/internal/storage"
	"github.com/mark3labs/mcp-go/server"
)

func TestResolveTargetUserID(t *testing.T) {
	tempDir := t.TempDir()
	store, err := storage.NewStore(filepath.Join(tempDir, "store.json"))
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	client := &discord.Client{OwnerUserID: "owner-from-client"}

	// 1. Argument takes top priority
	resolved := resolveTargetUserID("arg-user-123", client, store)
	if resolved != "arg-user-123" {
		t.Errorf("expected arg-user-123, got %s", resolved)
	}

	// 2. Client config fallback
	resolved = resolveTargetUserID("", client, store)
	if resolved != "owner-from-client" {
		t.Errorf("expected owner-from-client, got %s", resolved)
	}

	// 3. Persistent store fallback when client config is empty
	client.OwnerUserID = ""
	_ = store.Save(storage.CustomResource{
		Key:     "owner_user_id",
		Content: "owner-from-store",
	})
	resolved = resolveTargetUserID("", client, store)
	if resolved != "owner-from-store" {
		t.Errorf("expected owner-from-store, got %s", resolved)
	}

	// 4. Empty fallback
	_ = store.Delete("owner_user_id")
	resolved = resolveTargetUserID("", client, store)
	if resolved != "" {
		t.Errorf("expected empty string, got %s", resolved)
	}
}

func TestAskUserToolErrors(t *testing.T) {
	s := server.NewMCPServer("test-server", "1.0.0")
	store, _ := storage.NewStore("")
	RegisterCollaborationTools(s, nil, store)

	// Call without userId or owner configured
	req := []byte(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "ask_user",
			"arguments": {
				"question": "Can I deploy to production?"
			}
		}
	}`)

	resp := s.HandleMessage(context.Background(), req)
	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	rawResp, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var callResp struct {
		Result struct {
			IsError bool `json:"isError"`
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"result"`
	}
	if err := json.Unmarshal(rawResp, &callResp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !callResp.Result.IsError {
		t.Fatal("expected error response when user ID is unresolved")
	}
}
