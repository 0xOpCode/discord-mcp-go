package mcp

import (
	"context"
	"encoding/json"
	"os"
	"strings"
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

	toolNames := make(map[string]bool)
	for _, tool := range listResp.Result.Tools {
		toolNames[tool.Name] = true
	}

	expectedTools := []string{
		"run_pipeline",
		"get_message",
		"send_file",
		"send_private_file",
		"create_thread_from_message",
		"list_guild_members",
		"search_guild_messages",
		"set_event_image",
		"list_pinned_messages",
		"pin_message",
		"unpin_message",
		"crosspost_message",
		"save_custom_resource",
		"get_custom_resource",
		"delete_custom_resource",
		"list_custom_resources",
		"ask_user",
	}

	for _, name := range expectedTools {
		if !toolNames[name] {
			t.Fatalf("expected tool '%s' not found in registered tools", name)
		}
	}

	if toolCount != 113 {
		t.Fatalf("expected 113 tools, got %d", toolCount)
	}
}

func TestPipelineInterpolation(t *testing.T) {
	ctxMap := map[string]StepOutput{
		"step1": {
			ID:          "step1",
			Tool:        "create_category",
			Success:     true,
			Output:      "Category `Dev` created. ID: `154000111222333444`",
			ExtractedID: "154000111222333444",
		},
	}

	input := "Parent category is {{step1.id}} with status {{step1.status}}"
	result := interpolateString(input, ctxMap)
	expected := "Parent category is 154000111222333444 with status success"
	if result != expected {
		t.Fatalf("unexpected interpolation result: got '%s', want '%s'", result, expected)
	}

	dollarInput := "Channel parent is $step1.id"
	dollarResult := interpolateString(dollarInput, ctxMap)
	dollarExpected := "Channel parent is 154000111222333444"
	if dollarResult != dollarExpected {
		t.Fatalf("unexpected dollar interpolation: got '%s', want '%s'", dollarResult, dollarExpected)
	}
}

func TestRunPipelineRecursiveDisallowed(t *testing.T) {
	client := &discord.Client{}
	s := NewServer(client)

	stepsJSON := `[{"id":"bad","tool":"run_pipeline","arguments":{}}]`
	rawReq := []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"run_pipeline","arguments":{"steps":` + stepsJSON + `}}}`)

	resp := s.MCPServer.HandleMessage(context.Background(), rawReq)
	respBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	if !strings.Contains(string(respBytes), "recursive run_pipeline invocation is disallowed") {
		t.Fatalf("expected recursion rejection, got: %s", string(respBytes))
	}
}

func TestReadFileInput(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "testfile-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	content := "hello discord file test"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	data, name, err := readFileInput(tmpFile.Name(), "", nil)
	if err != nil {
		t.Fatalf("readFileInput failed: %v", err)
	}
	if string(data) != content {
		t.Fatalf("content mismatch: got %s, want %s", string(data), content)
	}
	if name == "" {
		t.Fatal("expected non-empty filename")
	}

	dataURI := "data:text/plain;base64,aGVsbG8="
	uriData, uriName, err := readFileInput(dataURI, "", nil)
	if err != nil {
		t.Fatalf("data URI decode failed: %v", err)
	}
	if string(uriData) != "hello" {
		t.Fatalf("data URI content mismatch: got %s, want hello", string(uriData))
	}
	if uriName != "attachment.txt" {
		t.Fatalf("expected attachment.txt, got %s", uriName)
	}

	allowedDir := os.TempDir()
	_, _, errAllowed := readFileInput(tmpFile.Name(), "", []string{allowedDir})
	if errAllowed != nil {
		t.Fatalf("expected file in %s to be allowed, got: %v", allowedDir, errAllowed)
	}

	_, _, errForbidden := readFileInput(tmpFile.Name(), "", []string{"/var/restricted/uploads"})
	if errForbidden == nil {
		t.Fatal("expected file outside allowed directories to be rejected")
	}

	_, _, errTraversal := readFileInput(tmpFile.Name()+"/../../etc/passwd", "", []string{allowedDir})
	if errTraversal == nil {
		t.Fatal("expected path traversal to be rejected")
	}

	uriDataRestricted, _, errUriRestricted := readFileInput(dataURI, "", []string{"/var/restricted/uploads"})
	if errUriRestricted != nil || string(uriDataRestricted) != "hello" {
		t.Fatalf("expected data URI to work with path restriction: %v", errUriRestricted)
	}
}
