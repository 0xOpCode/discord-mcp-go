package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreOperations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "store-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "resources.json")
	store, err := NewStore(filePath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Test Save
	item := CustomResource{
		Key:         "owner_user_id",
		Content:     "123456789012345678",
		Description: "Primary Discord User ID of server owner",
	}
	if err := store.Save(item); err != nil {
		t.Fatalf("failed to save item: %v", err)
	}

	// Test Get
	retrieved, err := store.Get("owner_user_id")
	if err != nil {
		t.Fatalf("failed to get item: %v", err)
	}
	if retrieved.Content != item.Content {
		t.Fatalf("content mismatch: got %s, want %s", retrieved.Content, item.Content)
	}
	if retrieved.URI != "discord://memory/owner_user_id" {
		t.Fatalf("uri mismatch: got %s", retrieved.URI)
	}

	// Test Get with URI prefix
	retrievedByURI, err := store.Get("discord://memory/owner_user_id")
	if err != nil {
		t.Fatalf("failed to get item by URI: %v", err)
	}
	if retrievedByURI.Key != "owner_user_id" {
		t.Fatalf("key mismatch: got %s", retrievedByURI.Key)
	}

	// Test List
	list := store.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 item, got %d", len(list))
	}

	// Test Persistence across store instances
	storeReopened, err := NewStore(filePath)
	if err != nil {
		t.Fatalf("failed to reopen store: %v", err)
	}
	reopenedItem, err := storeReopened.Get("owner_user_id")
	if err != nil {
		t.Fatalf("failed to get item from reopened store: %v", err)
	}
	if reopenedItem.Content != item.Content {
		t.Fatalf("persistence mismatch: got %s, want %s", reopenedItem.Content, item.Content)
	}

	// Test Delete
	if err := store.Delete("owner_user_id"); err != nil {
		t.Fatalf("failed to delete item: %v", err)
	}
	_, errNotFound := store.Get("owner_user_id")
	if errNotFound == nil {
		t.Fatal("expected error after delete, got nil")
	}
}
