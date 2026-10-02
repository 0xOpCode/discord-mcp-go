package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/google/uuid"
)

type sseSession struct {
	writer  http.ResponseWriter
	flusher http.Flusher
	done    chan struct{}
}

type HTTPServer struct {
	server   *Server
	baseURL  string
	sessions sync.Map
}

func NewHTTPServer(server *Server, baseURL string) *HTTPServer {
	return &HTTPServer{
		server:  server,
		baseURL: baseURL,
	}
}

func (h *HTTPServer) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/sse", h.handleSSEOrPost)
	mux.HandleFunc("/message", h.handleMessage)
	mux.HandleFunc("/mcp", h.handleSSEOrPost)
	mux.HandleFunc("/", h.handleRoot)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		mux.ServeHTTP(w, r)
	})
}

func (h *HTTPServer) handleSSEOrPost(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.handleDirectJSONRPC(w, r)
		return
	}
	if r.Method == http.MethodGet {
		h.handleSSE(w, r)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (h *HTTPServer) handleDirectJSONRPC(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read body: %v", err), http.StatusBadRequest)
		return
	}

	response := h.server.MCPServer.HandleMessage(r.Context(), body)
	w.Header().Set("Content-Type", "application/json")
	if response != nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	} else {
		w.WriteHeader(http.StatusAccepted)
	}
}

func (h *HTTPServer) handleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		h.handleDirectJSONRPC(w, r)
		return
	}

	sessionI, ok := h.sessions.Load(sessionID)
	if !ok {
		h.handleDirectJSONRPC(w, r)
		return
	}
	session := sessionI.(*sseSession)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Parse error", http.StatusBadRequest)
		return
	}

	response := h.server.MCPServer.HandleMessage(r.Context(), body)
	if response != nil {
		eventData, _ := json.Marshal(response)
		fmt.Fprintf(session.writer, "event: message\ndata: %s\n\n", eventData)
		session.flusher.Flush()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(response)
	} else {
		w.WriteHeader(http.StatusAccepted)
	}
}

func (h *HTTPServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	sessionID := uuid.New().String()
	session := &sseSession{
		writer:  w,
		flusher: flusher,
		done:    make(chan struct{}),
	}

	h.sessions.Store(sessionID, session)
	defer h.sessions.Delete(sessionID)

	messageEndpoint := fmt.Sprintf("%s/message?sessionId=%s", h.baseURL, sessionID)
	fmt.Fprintf(w, "event: endpoint\ndata: %s\r\n\r\n", messageEndpoint)
	flusher.Flush()

	<-r.Context().Done()
	close(session.done)
}

func (h *HTTPServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.handleDirectJSONRPC(w, r)
		return
	}
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","server":"discord-mcp-go","version":"1.2.1"}`))
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
