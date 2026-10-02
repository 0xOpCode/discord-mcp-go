package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/0xOpCode/discord-mcp-go/internal/discord"
	"github.com/bwmarrin/discordgo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterRoleConnectionTools(s *server.MCPServer, client *discord.Client) {
	// get_role_connection
	s.AddTool(
		mcp.NewTool("get_role_connection",
			mcp.WithDescription("Get the application role connection for the bot user"),
			mcp.WithString("applicationId", mcp.Description("Optional Discord Application ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			appID, err := resolveAppID(client, getString(req.Params.Arguments, "applicationId"))
			if err != nil {
				return errorResult(err), nil
			}

			rconn, err := client.Session.UserApplicationRoleConnection(appID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to fetch user role connection: %w", err)), nil
			}

			if rconn == nil {
				return successResult(fmt.Sprintf("No role connection found for application `%s`.", appID)), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Application Role Connection (App: `%s`)\n\n", appID))
			sb.WriteString(fmt.Sprintf("- **Platform Name:** %s\n", rconn.PlatformName))
			sb.WriteString(fmt.Sprintf("- **Platform Username:** %s\n", rconn.PlatformUsername))
			sb.WriteString(fmt.Sprintf("- **Metadata Count:** %d\n\n", len(rconn.Metadata)))

			if len(rconn.Metadata) > 0 {
				sb.WriteString("| Key | Value |\n|---|---|\n")
				for k, v := range rconn.Metadata {
					sb.WriteString(fmt.Sprintf("| `%s` | `%s` |\n", k, v))
				}
			}

			return successResult(sb.String()), nil
		},
	)

	// update_role_connection
	s.AddTool(
		mcp.NewTool("update_role_connection",
			mcp.WithDescription("Update the application role connection for the bot user"),
			mcp.WithString("applicationId", mcp.Description("Optional Discord Application ID")),
			mcp.WithString("platformName", mcp.Description("Platform name (max 50 chars)")),
			mcp.WithString("platformUsername", mcp.Description("Platform username (max 100 chars)")),
			mcp.WithString("metadata", mcp.Description("JSON map of string keys to string values")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			appID, err := resolveAppID(client, getString(req.Params.Arguments, "applicationId"))
			if err != nil {
				return errorResult(err), nil
			}

			rconn := &discordgo.ApplicationRoleConnection{
				PlatformName:     getString(req.Params.Arguments, "platformName"),
				PlatformUsername: getString(req.Params.Arguments, "platformUsername"),
				Metadata:         make(map[string]string),
			}

			metaRaw := getString(req.Params.Arguments, "metadata")
			if metaRaw != "" {
				if err := json.Unmarshal([]byte(metaRaw), &rconn.Metadata); err != nil {
					return errorResult(fmt.Errorf("invalid metadata JSON: %w", err)), nil
				}
			}

			updated, err := client.Session.UserApplicationRoleConnectionUpdate(appID, rconn)
			if err != nil {
				return errorResult(fmt.Errorf("failed to update user role connection: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Role connection updated for application `%s`. Platform: `%s` (%s)", appID, updated.PlatformName, updated.PlatformUsername)), nil
		},
	)

	// get_role_connection_metadata
	s.AddTool(
		mcp.NewTool("get_role_connection_metadata",
			mcp.WithDescription("Get application role connection metadata records for linked roles"),
			mcp.WithString("applicationId", mcp.Description("Optional Discord Application ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			appID, err := resolveAppID(client, getString(req.Params.Arguments, "applicationId"))
			if err != nil {
				return errorResult(err), nil
			}

			records, err := client.Session.ApplicationRoleConnectionMetadata(appID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to fetch role connection metadata: %w", err)), nil
			}

			if len(records) == 0 {
				return successResult(fmt.Sprintf("No role connection metadata records found for application `%s`.", appID)), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Application Role Connection Metadata (App: `%s`, %d records)\n\n", appID, len(records)))
			sb.WriteString("| Key | Name | Type | Description |\n|---|---|---|---|\n")
			for _, rec := range records {
				sb.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s |\n", rec.Key, rec.Name, metadataTypeName(rec.Type), rec.Description))
			}

			return successResult(sb.String()), nil
		},
	)

	// update_role_connection_metadata
	s.AddTool(
		mcp.NewTool("update_role_connection_metadata",
			mcp.WithDescription("Update application role connection metadata records for linked roles"),
			mcp.WithString("records", mcp.Required(), mcp.Description("JSON array of metadata records (key, name, description, type)")),
			mcp.WithString("applicationId", mcp.Description("Optional Discord Application ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			appID, err := resolveAppID(client, getString(req.Params.Arguments, "applicationId"))
			if err != nil {
				return errorResult(err), nil
			}

			recordsRaw := getString(req.Params.Arguments, "records")
			var records []*discordgo.ApplicationRoleConnectionMetadata
			if err := json.Unmarshal([]byte(recordsRaw), &records); err != nil {
				return errorResult(fmt.Errorf("invalid records JSON: %w", err)), nil
			}

			updated, err := client.Session.ApplicationRoleConnectionMetadataUpdate(appID, records)
			if err != nil {
				return errorResult(fmt.Errorf("failed to update role connection metadata: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Updated %d role connection metadata records for application `%s`.", len(updated), appID)), nil
		},
	)
}

func resolveAppID(client *discord.Client, appID string) (string, error) {
	if appID != "" {
		return appID, nil
	}
	if client.Session.State != nil && client.Session.State.User != nil && client.Session.State.User.ID != "" {
		return client.Session.State.User.ID, nil
	}
	u, err := client.Session.User("@me")
	if err != nil {
		return "", fmt.Errorf("unable to resolve application ID: %w", err)
	}
	return u.ID, nil
}

func metadataTypeName(t discordgo.ApplicationRoleConnectionMetadataType) string {
	switch t {
	case discordgo.ApplicationRoleConnectionMetadataIntegerLessThanOrEqual:
		return "IntegerLessThanOrEqual"
	case discordgo.ApplicationRoleConnectionMetadataIntegerGreaterThanOrEqual:
		return "IntegerGreaterThanOrEqual"
	case discordgo.ApplicationRoleConnectionMetadataIntegerEqual:
		return "IntegerEqual"
	case discordgo.ApplicationRoleConnectionMetadataIntegerNotEqual:
		return "IntegerNotEqual"
	case discordgo.ApplicationRoleConnectionMetadataDatetimeLessThanOrEqual:
		return "DatetimeLessThanOrEqual"
	case discordgo.ApplicationRoleConnectionMetadataDatetimeGreaterThanOrEqual:
		return "DatetimeGreaterThanOrEqual"
	case discordgo.ApplicationRoleConnectionMetadataBooleanEqual:
		return "BooleanEqual"
	case discordgo.ApplicationRoleConnectionMetadataBooleanNotEqual:
		return "BooleanNotEqual"
	default:
		return "Unknown"
	}
}
