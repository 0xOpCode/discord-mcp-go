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
	}

	for _, name := range expectedTools {
		if !toolNames[name] {
			t.Fatalf("expected tool '%s' not found in registered tools", name)
		}
	}

	if toolCount != 100 {
		t.Fatalf("expected 100 tools, got %d", toolCount)
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

	data, name, err := readFileInput(tmpFile.Name(), "")
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
	uriData, uriName, err := readFileInput(dataURI, "")
	if err != nil {
		t.Fatalf("data URI decode failed: %v", err)
	}
	if string(uriData) != "hello" {
		t.Fatalf("data URI content mismatch: got %s, want hello", string(uriData))
	}
	if uriName != "attachment.txt" {
		t.Fatalf("expected attachment.txt, got %s", uriName)
	}
}
