package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/0xOpCode/discord-mcp-go/internal/discord"
)

func TestRegisteredToolsCount(t *testing.T) {
	client := &discord.Client{}
	s := NewServer(client)

	rawReq := []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
	resp := s.MCPServer.HandleMessage(context.Background(), rawReq)

	respJSON, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var listResp struct {
		Result struct {
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(respJSON, &listResp); err != nil {
		t.Fatalf("failed to unmarshal tools/list response: %v", err)
	}

	toolCount := len(listResp.Result.Tools)
	t.Logf("Total registered tools: %d", toolCount)
	for i, tool := range listResp.Result.Tools {
		t.Logf("  [%d] %s", i+1, tool.Name)
	}

	if toolCount < 90 {
		t.Fatalf("expected at least 90 tools, got %d", toolCount)
	}
}
