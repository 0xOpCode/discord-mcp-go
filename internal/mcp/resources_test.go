package mcp

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/0xOpCode/discord-mcp-go/internal/storage"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestMCPResources(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "resources.json")
	store, err := storage.NewStore(storePath)
	if err != nil {
		t.Fatalf("failed to init store: %v", err)
	}

	err = store.Save(storage.CustomResource{
		Key:         "owner_guide",
		Content:     "Contact owner on discord @admin",
		Description: "Guidelines for bot interaction",
		MimeType:    "text/plain",
	})
	if err != nil {
		t.Fatalf("failed to save resource: %v", err)
	}

	s := server.NewMCPServer("test-server", "1.0.0", server.WithResourceCapabilities(true, true))
	RegisterResources(s, nil, store)

	// Test 1: resources/templates/list
	templatesReq := []byte(`{"jsonrpc":"2.0","id":1,"method":"resources/templates/list","params":{}}`)
	resp := s.HandleMessage(context.Background(), templatesReq)
	if resp == nil {
		t.Fatal("expected response for templates/list, got nil")
	}

	rawResp, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var templatesResult struct {
		Result struct {
			ResourceTemplates []mcp.ResourceTemplate `json:"resourceTemplates"`
		} `json:"result"`
	}
	if err := json.Unmarshal(rawResp, &templatesResult); err != nil {
		t.Fatalf("failed to unmarshal templates response: %v", err)
	}

	expectedTemplates := map[string]bool{
		"discord://memory/{key}":              false,
		"discord://guilds/{guildId}/overview": false,
		"discord://channels/{channelId}/pinned": false,
		"discord://channels/{channelId}/recent": false,
		"discord://guilds/{guildId}/roles":    false,
	}

	for _, tmpl := range templatesResult.Result.ResourceTemplates {
		if _, ok := expectedTemplates[tmpl.URITemplate]; ok {
			expectedTemplates[tmpl.URITemplate] = true
		}
	}

	for tmpl, found := range expectedTemplates {
		if !found {
			t.Errorf("expected template %s not found", tmpl)
		}
	}

	// Test 2: resources/list contains persistent memory resource
	listReq := []byte(`{"jsonrpc":"2.0","id":2,"method":"resources/list","params":{}}`)
	resp = s.HandleMessage(context.Background(), listReq)
	if resp == nil {
		t.Fatal("expected response for resources/list, got nil")
	}

	rawResp, err = json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal list response: %v", err)
	}

	var listResult struct {
		Result struct {
			Resources []mcp.Resource `json:"resources"`
		} `json:"result"`
	}
	if err := json.Unmarshal(rawResp, &listResult); err != nil {
		t.Fatalf("failed to unmarshal list response: %v", err)
	}

	foundMemoryResource := false
	for _, res := range listResult.Result.Resources {
		if res.URI == "discord://memory/owner_guide" {
			foundMemoryResource = true
			break
		}
	}
	if !foundMemoryResource {
		t.Error("expected discord://memory/owner_guide to be listed in resources/list")
	}

	// Test 3: resources/read for discord://memory/owner_guide
	readReq := []byte(`{"jsonrpc":"2.0","id":3,"method":"resources/read","params":{"uri":"discord://memory/owner_guide"}}`)
	resp = s.HandleMessage(context.Background(), readReq)
	if resp == nil {
		t.Fatal("expected response for resources/read, got nil")
	}

	rawResp, err = json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal read response: %v", err)
	}

	var readResult struct {
		Result struct {
			Contents []struct {
				URI  string `json:"uri"`
				Text string `json:"text"`
			} `json:"contents"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rawResp, &readResult); err != nil {
		t.Fatalf("failed to unmarshal read response: %v", err)
	}

	if readResult.Error != nil {
		t.Fatalf("unexpected error reading resource: %v", readResult.Error.Message)
	}

	if len(readResult.Result.Contents) != 1 {
		t.Fatalf("expected 1 content entry, got %d", len(readResult.Result.Contents))
	}

	if readResult.Result.Contents[0].Text != "Contact owner on discord @admin" {
		t.Errorf("unexpected content text: %s", readResult.Result.Contents[0].Text)
	}

	// Test 4: verify embedded guides in resources/list
	expectedGuides := map[string]bool{
		"discord://guide/overview":       false,
		"discord://guide/tools":          false,
		"discord://guide/pipelines":      false,
		"discord://guide/best-practices": false,
	}

	for _, res := range listResult.Result.Resources {
		if _, ok := expectedGuides[res.URI]; ok {
			expectedGuides[res.URI] = true
		}
	}

	for uri, found := range expectedGuides {
		if !found {
			t.Errorf("expected guide resource %s not found in resources/list", uri)
		}
	}

	// Test 5: resources/read for discord://guide/pipelines
	guideReq := []byte(`{"jsonrpc":"2.0","id":4,"method":"resources/read","params":{"uri":"discord://guide/pipelines"}}`)
	resp = s.HandleMessage(context.Background(), guideReq)
	if resp == nil {
		t.Fatal("expected response for guide read, got nil")
	}

	rawResp, err = json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal guide response: %v", err)
	}

	var guideResult struct {
		Result struct {
			Contents []struct {
				URI      string `json:"uri"`
				MIMEType string `json:"mimeType"`
				Text     string `json:"text"`
			} `json:"contents"`
		} `json:"result"`
	}
	if err := json.Unmarshal(rawResp, &guideResult); err != nil {
		t.Fatalf("failed to unmarshal guide response: %v", err)
	}

	if len(guideResult.Result.Contents) != 1 {
		t.Fatalf("expected 1 content entry for guide, got %d", len(guideResult.Result.Contents))
	}

	if guideResult.Result.Contents[0].MIMEType != "text/markdown" {
		t.Errorf("expected text/markdown mimeType, got %s", guideResult.Result.Contents[0].MIMEType)
	}
}
