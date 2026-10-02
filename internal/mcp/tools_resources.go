package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/0xOpCode/discord-mcp-go/internal/storage"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterCustomResourceTools(s *server.MCPServer, store *storage.Store) {
	// save_custom_resource
	s.AddTool(
		mcp.NewTool("save_custom_resource",
			mcp.WithDescription("Save a persistent resource document (notes, runbook, guidelines, or owner config) under a custom key"),
			mcp.WithString("key", mcp.Required(), mcp.Description("Unique resource identifier (e.g. owner_user_id, project_context)")),
			mcp.WithString("content", mcp.Required(), mcp.Description("Resource payload content (markdown, plain text, or JSON)")),
			mcp.WithString("description", mcp.Description("Optional human-readable summary of the resource")),
			mcp.WithString("mimeType", mcp.Description("MIME type (default: text/plain)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			key := getString(req.Params.Arguments, "key")
			content := getString(req.Params.Arguments, "content")
			if key == "" || content == "" {
				return errorResult(fmt.Errorf("key and content are required")), nil
			}

			item := storage.CustomResource{
				Key:         key,
				Content:     content,
				Description: getString(req.Params.Arguments, "description"),
				MimeType:    getString(req.Params.Arguments, "mimeType"),
			}

			if err := store.Save(item); err != nil {
				return errorResult(fmt.Errorf("failed to save resource: %w", err)), nil
			}
			return successResult(fmt.Sprintf("Resource `%s` saved. URI: `discord://memory/%s`", key, key)), nil
		},
	)

	// get_custom_resource
	s.AddTool(
		mcp.NewTool("get_custom_resource",
			mcp.WithDescription("Retrieve a persistent custom resource document by key or URI"),
			mcp.WithString("key", mcp.Required(), mcp.Description("Resource key (e.g. owner_user_id or discord://memory/owner_user_id)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			key := getString(req.Params.Arguments, "key")
			if key == "" {
				return errorResult(fmt.Errorf("key is required")), nil
			}

			item, err := store.Get(key)
			if err != nil {
				return errorResult(err), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Resource: `%s` (`%s`)\n", item.Key, item.URI))
			if item.Description != "" {
				sb.WriteString(fmt.Sprintf("**Description**: %s\n", item.Description))
			}
			sb.WriteString(fmt.Sprintf("**Updated**: %s | **MIME**: %s\n\n", item.UpdatedAt.Format("2006-01-02 15:04:05 UTC"), item.MimeType))
			sb.WriteString(item.Content)

			return successResult(sb.String()), nil
		},
	)

	// delete_custom_resource
	s.AddTool(
		mcp.NewTool("delete_custom_resource",
			mcp.WithDescription("Delete a persistent custom resource document by key or URI"),
			mcp.WithString("key", mcp.Required(), mcp.Description("Resource key or URI to delete")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			key := getString(req.Params.Arguments, "key")
			if key == "" {
				return errorResult(fmt.Errorf("key is required")), nil
			}

			if err := store.Delete(key); err != nil {
				return errorResult(err), nil
			}
			return successResult(fmt.Sprintf("Resource `%s` deleted.", key)), nil
		},
	)

	// list_custom_resources
	s.AddTool(
		mcp.NewTool("list_custom_resources",
			mcp.WithDescription("List all persistent custom resources stored on disk"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			items := store.List()
			if len(items) == 0 {
				return successResult("No custom resources stored."), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Custom Persistent Resources (%d total):\n\n", len(items)))
			sb.WriteString(fmt.Sprintf("| %-22s | %-32s | %-20s | %-25s |\n", "Key", "URI", "Updated", "Description"))
			sb.WriteString("|------------------------|----------------------------------|----------------------|---------------------------|\n")

			for _, item := range items {
				desc := item.Description
				if len(desc) > 25 {
					desc = desc[:22] + "..."
				}
				updated := item.UpdatedAt.Format("2006-01-02 15:04")
				sb.WriteString(fmt.Sprintf("| %-22s | %-32s | %-20s | %-25s |\n", item.Key, item.URI, updated, desc))
			}

			return successResult(sb.String()), nil
		},
	)
}
