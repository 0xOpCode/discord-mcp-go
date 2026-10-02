package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/0xOpCode/discord-mcp-go/internal/discord"
	"github.com/bwmarrin/discordgo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterServerTools(s *server.MCPServer, client *discord.Client) {
	// list_servers
	s.AddTool(
		mcp.NewTool("list_servers",
			mcp.WithDescription("List all Discord servers (guilds) the bot is currently in, including name, ID, and owner status"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			guilds, err := client.ListServers()
			if err != nil {
				return errorResult(err), nil
			}
			if len(guilds) == 0 {
				return successResult("Bot is not in any Discord servers."), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Connected Discord Servers (%d total):\n\n", len(guilds)))
			sb.WriteString(fmt.Sprintf("| %-28s | %-20s | %-8s |\n", "Server Name", "Server ID", "Owner"))
			sb.WriteString("|------------------------------|----------------------|----------|\n")
			for _, g := range guilds {
				name := g.Name
				if len(name) > 28 {
					name = name[:25] + "..."
				}
				sb.WriteString(fmt.Sprintf("| %-28s | %-20s | %-8t |\n", name, g.ID, g.Owner))
			}
			return successResult(sb.String()), nil
		},
	)

	// get_server_info
	s.AddTool(
		mcp.NewTool("get_server_info",
			mcp.WithDescription("Get detailed information about a Discord server"),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server (Guild) ID. Uses default if omitted.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			guildID := getString(req.Params.Arguments, "guildId")
			guild, err := client.GetServerInfo(guildID)
			if err != nil {
				return errorResult(err), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Server Information: %s\n", guild.Name))
			sb.WriteString(fmt.Sprintf("- **ID**: `%s`\n", guild.ID))
			sb.WriteString(fmt.Sprintf("- **Owner ID**: `<@%s>`\n", guild.OwnerID))
			sb.WriteString(fmt.Sprintf("- **Member Count**: %d\n", guild.MemberCount))
			sb.WriteString(fmt.Sprintf("- **Channels**: %d\n", len(guild.Channels)))
			sb.WriteString(fmt.Sprintf("- **Roles**: %d\n", len(guild.Roles)))
			sb.WriteString(fmt.Sprintf("- **Emojis**: %d\n", len(guild.Emojis)))
			sb.WriteString(fmt.Sprintf("- **AFK Timeout**: %d seconds\n", guild.AfkTimeout))
			if guild.Description != "" {
				sb.WriteString(fmt.Sprintf("- **Description**: %s\n", guild.Description))
			}
			return successResult(sb.String()), nil
		},
	)

	// check_bot_permissions
	s.AddTool(
		mcp.NewTool("check_bot_permissions",
			mcp.WithDescription("Check permissions the bot holds in a specific Discord server"),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server (Guild) ID. Uses default if omitted.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			guildID := getString(req.Params.Arguments, "guildId")
			perms, err := client.CheckBotPermissions(guildID)
			if err != nil {
				return errorResult(err), nil
			}

			var sb strings.Builder
			sb.WriteString("### Bot Permissions Breakdown:\n\n")
			sb.WriteString(fmt.Sprintf("| %-25s | %-8s |\n", "Permission", "Granted"))
			sb.WriteString("|---------------------------|----------|\n")
			for perm, granted := range perms {
				status := "NO"
				if granted {
					status = "YES"
				}
				sb.WriteString(fmt.Sprintf("| %-25s | %-8s |\n", perm, status))
			}
			return successResult(sb.String()), nil
		},
	)

	// get_audit_logs
	s.AddTool(
		mcp.NewTool("get_audit_logs",
			mcp.WithDescription("Inspect recent audit log entries in a Discord server"),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
			mcp.WithString("userId", mcp.Description("Optional user ID filter")),
			mcp.WithNumber("limit", mcp.Description("Number of entries to retrieve (1-100, default 50)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			limit := getInt(req.Params.Arguments, "limit", 50)
			if limit < 1 {
				limit = 1
			} else if limit > 100 {
				limit = 100
			}

			userID := getString(req.Params.Arguments, "userId")
			auditLog, err := client.Session.GuildAuditLog(guildID, userID, "", 0, limit)
			if err != nil {
				return errorResult(fmt.Errorf("failed to retrieve audit log: %w", err)), nil
			}

			if len(auditLog.AuditLogEntries) == 0 {
				return successResult("No audit log entries found."), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Audit Log for Server `%s` (%d entries):\n\n", guildID, len(auditLog.AuditLogEntries)))
			sb.WriteString(fmt.Sprintf("| %-20s | %-20s | %-12s | %-25s |\n", "Actor User ID", "Target ID", "Action Type", "Reason"))
			sb.WriteString("|----------------------|----------------------|--------------|---------------------------|\n")
			for _, e := range auditLog.AuditLogEntries {
				reason := e.Reason
				if reason == "" {
					reason = "None provided"
				}
				if len(reason) > 25 {
					reason = reason[:22] + "..."
				}
				sb.WriteString(fmt.Sprintf("| %-20s | %-20s | %-12d | %-25s |\n", e.UserID, e.TargetID, e.ActionType, reason))
			}
			return successResult(sb.String()), nil
		},
	)

	// set_bot_activity
	s.AddTool(
		mcp.NewTool("set_bot_activity",
			mcp.WithDescription("Set the Discord bot's activity presence and online status"),
			mcp.WithString("activityName", mcp.Required(), mcp.Description("Activity name (e.g. 'with AI', 'over servers')")),
			mcp.WithString("activityType", mcp.Description("Type: 'playing', 'streaming', 'listening', 'watching', 'competing' (default: playing)")),
			mcp.WithString("status", mcp.Description("Status: 'online', 'idle', 'dnd', 'invisible' (default: online)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name := getString(req.Params.Arguments, "activityName")
			if name == "" {
				return errorResult(fmt.Errorf("activityName is required")), nil
			}

			actTypeStr := strings.ToLower(getString(req.Params.Arguments, "activityType"))
			var actType discordgo.ActivityType = discordgo.ActivityTypeGame
			switch actTypeStr {
			case "streaming":
				actType = discordgo.ActivityTypeStreaming
			case "listening":
				actType = discordgo.ActivityTypeListening
			case "watching":
				actType = discordgo.ActivityTypeWatching
			case "competing":
				actType = discordgo.ActivityTypeCompeting
			}

			status := strings.ToLower(getString(req.Params.Arguments, "status"))
			if status == "" {
				status = "online"
			}

			data := discordgo.UpdateStatusData{
				Status: status,
				Activities: []*discordgo.Activity{
					{
						Name: name,
						Type: actType,
					},
				},
			}

			err := client.Session.UpdateStatusComplex(data)
			if err != nil {
				return errorResult(fmt.Errorf("failed to update status: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Bot status updated: %s (%s)", name, status)), nil
		},
	)
}
