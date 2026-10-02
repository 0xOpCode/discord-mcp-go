package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/0xOpCode/discord-mcp-go/internal/discord"
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
}
