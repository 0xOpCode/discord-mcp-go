package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type CustomResource struct {
	Key         string    `json:"key"`
	URI         string    `json:"uri"`
	Name        string    `json:"name"`
	Content     string    `json:"content"`
	Description string    `json:"description"`
	MimeType    string    `json:"mimeType"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Store struct {
	filePath string
	mu       sync.RWMutex
	items    map[string]CustomResource
}

func ResolveDefaultPath() string {
	if dataDir := strings.TrimSpace(os.Getenv("DISCORD_DATA_DIR")); dataDir != "" {
		return filepath.Join(dataDir, "resources.json")
	}

	configDir, err := os.UserConfigDir()
	if err == nil && configDir != "" {
		return filepath.Join(configDir, "discord-mcp-go", "resources.json")
	}

	homeDir, err := os.UserHomeDir()
	if err == nil && homeDir != "" {
		return filepath.Join(homeDir, ".config", "discord-mcp-go", "resources.json")
	}

	return "./data/resources.json"
}

func NewStore(filePath string) (*Store, error) {
	if filePath == "" {
		filePath = ResolveDefaultPath()
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	s := &Store{
		filePath: filePath,
		items:    make(map[string]CustomResource),
	}

	if data, err := os.ReadFile(filePath); err == nil && len(data) > 0 {
		var loaded map[string]CustomResource
		if err := json.Unmarshal(data, &loaded); err == nil {
			s.items = loaded
		}
	}

	return s, nil
}

func (s *Store) Save(item CustomResource) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	item.Key = strings.TrimSpace(item.Key)
	if item.Key == "" {
		return fmt.Errorf("resource key required")
	}

	if item.URI == "" {
		item.URI = "discord://memory/" + item.Key
	}
	if item.Name == "" {
		item.Name = item.Key
	}
	if item.MimeType == "" {
		item.MimeType = "text/plain"
	}
	item.UpdatedAt = time.Now().UTC()

	s.items[item.Key] = item
	return s.persist()
}

func (s *Store) Get(key string) (*CustomResource, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cleanKey := strings.TrimSpace(key)
	cleanKey = strings.TrimPrefix(cleanKey, "discord://memory/")

	item, exists := s.items[cleanKey]
	if !exists {
		return nil, fmt.Errorf("custom resource '%s' not found", key)
	}
	return &item, nil
}

func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cleanKey := strings.TrimSpace(key)
	cleanKey = strings.TrimPrefix(cleanKey, "discord://memory/")

	if _, exists := s.items[cleanKey]; !exists {
		return fmt.Errorf("custom resource '%s' not found", key)
	}

	delete(s.items, cleanKey)
	return s.persist()
}

func (s *Store) List() []CustomResource {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]CustomResource, 0, len(s.items))
	for _, item := range s.items {
		list = append(list, item)
	}
	return list
}

func (s *Store) persist() error {
	data, err := json.MarshalIndent(s.items, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode storage items: %w", err)
	}

	tmpFile := s.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary storage file: %w", err)
	}

	if err := os.Rename(tmpFile, s.filePath); err != nil {
		return fmt.Errorf("failed to atomic commit storage file: %w", err)
	}

	return nil
}
