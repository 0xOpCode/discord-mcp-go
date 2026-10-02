package mcp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/0xOpCode/discord-mcp-go/internal/discord"
)

func TestHTTPServerDirectJSONRPC(t *testing.T) {
	client := &discord.Client{}
	s := NewServer(client)
	httpServer := NewHTTPServer(s, "http://localhost:8085")
	handler := httpServer.Handler()

	initReq := []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0"}}}`)

	// Test POST /sse (Streamable HTTP / direct JSON-RPC)
	reqSSE := httptest.NewRequest(http.MethodPost, "/sse", bytes.NewReader(initReq))
	reqSSE.Header.Set("Content-Type", "application/json")
	wSSE := httptest.NewRecorder()
	handler.ServeHTTP(wSSE, reqSSE)

	if wSSE.Code != http.StatusOK {
		t.Fatalf("expected status 200 for POST /sse, got %d: %s", wSSE.Code, wSSE.Body.String())
	}

	var initResp map[string]interface{}
	if err := json.Unmarshal(wSSE.Body.Bytes(), &initResp); err != nil {
		t.Fatalf("failed to decode JSON-RPC initialize response: %v", err)
	}
	if initResp["id"] == nil || initResp["result"] == nil {
		t.Fatalf("invalid initialize response payload: %s", wSSE.Body.String())
	}

	// Test POST /mcp
	reqMCP := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(initReq))
	reqMCP.Header.Set("Content-Type", "application/json")
	wMCP := httptest.NewRecorder()
	handler.ServeHTTP(wMCP, reqMCP)

	if wMCP.Code != http.StatusOK {
		t.Fatalf("expected status 200 for POST /mcp, got %d: %s", wMCP.Code, wMCP.Body.String())
	}

	// Test GET /
	reqRoot := httptest.NewRequest(http.MethodGet, "/", nil)
	wRoot := httptest.NewRecorder()
	handler.ServeHTTP(wRoot, reqRoot)
	if wRoot.Code != http.StatusOK || !strings.Contains(wRoot.Body.String(), "discord-mcp-go") {
		t.Fatalf("expected health status, got: %s", wRoot.Body.String())
	}
}

func TestHTTPServerSSEEndpoint(t *testing.T) {
	client := &discord.Client{}
	s := NewServer(client)
	httpServer := NewHTTPServer(s, "http://localhost:8085")
	ts := httptest.NewServer(httpServer.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/sse")
	if err != nil {
		t.Fatalf("failed to GET /sse: %v", err)
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/event-stream") {
		t.Fatalf("expected text/event-stream, got %s", contentType)
	}
}
