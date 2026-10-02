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

func RegisterPermissionTools(s *server.MCPServer, client *discord.Client) {
	// list_channel_permission_overwrites
	s.AddTool(
		mcp.NewTool("list_channel_permission_overwrites",
			mcp.WithDescription("List all permission overwrites configured on a Discord channel"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Channel ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			if channelID == "" {
				return errorResult(fmt.Errorf("channelId is required")), nil
			}

			ch, err := client.Session.Channel(channelID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to fetch channel: %w", err)), nil
			}

			if len(ch.PermissionOverwrites) == 0 {
				return successResult(fmt.Sprintf("No permission overwrites found on channel #%s.", ch.Name)), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Permission Overwrites for #%s (%d total):\n\n", ch.Name, len(ch.PermissionOverwrites)))
			sb.WriteString(fmt.Sprintf("| %-20s | %-10s | %-16s | %-16s |\n", "Target ID", "Type", "Allow Bits", "Deny Bits"))
			sb.WriteString("|----------------------|------------|------------------|------------------|\n")
			for _, po := range ch.PermissionOverwrites {
				t := "Role"
				if po.Type == discordgo.PermissionOverwriteTypeMember {
					t = "Member"
				}
				sb.WriteString(fmt.Sprintf("| %-20s | %-10s | 0x%014X | 0x%014X |\n", po.ID, t, po.Allow, po.Deny))
			}
			return successResult(sb.String()), nil
		},
	)

	// upsert_role_channel_permissions
	s.AddTool(
		mcp.NewTool("upsert_role_channel_permissions",
			mcp.WithDescription("Create or update permission overwrite for a role on a channel"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Channel ID")),
			mcp.WithString("roleId", mcp.Required(), mcp.Description("Role ID")),
			mcp.WithNumber("allow", mcp.Description("Allowed permissions bitmask")),
			mcp.WithNumber("deny", mcp.Description("Denied permissions bitmask")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			roleID := getString(req.Params.Arguments, "roleId")
			if channelID == "" || roleID == "" {
				return errorResult(fmt.Errorf("channelId and roleId are required")), nil
			}

			allow := int64(getInt(req.Params.Arguments, "allow", 0))
			deny := int64(getInt(req.Params.Arguments, "deny", 0))

			err := client.Session.ChannelPermissionSet(channelID, roleID, discordgo.PermissionOverwriteTypeRole, allow, deny)
			if err != nil {
				return errorResult(fmt.Errorf("failed to set role permission overwrite: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Permission overwrite for role `%s` updated on channel `%s`.", roleID, channelID)), nil
		},
	)

	// upsert_member_channel_permissions
	s.AddTool(
		mcp.NewTool("upsert_member_channel_permissions",
			mcp.WithDescription("Create or update permission overwrite for a member on a channel"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Channel ID")),
			mcp.WithString("memberId", mcp.Required(), mcp.Description("Member (User) ID")),
			mcp.WithNumber("allow", mcp.Description("Allowed permissions bitmask")),
			mcp.WithNumber("deny", mcp.Description("Denied permissions bitmask")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			memberID := getString(req.Params.Arguments, "memberId")
			if channelID == "" || memberID == "" {
				return errorResult(fmt.Errorf("channelId and memberId are required")), nil
			}

			allow := int64(getInt(req.Params.Arguments, "allow", 0))
			deny := int64(getInt(req.Params.Arguments, "deny", 0))

			err := client.Session.ChannelPermissionSet(channelID, memberID, discordgo.PermissionOverwriteTypeMember, allow, deny)
			if err != nil {
				return errorResult(fmt.Errorf("failed to set member permission overwrite: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Permission overwrite for member `%s` updated on channel `%s`.", memberID, channelID)), nil
		},
	)

	// delete_channel_permission_overwrite
	s.AddTool(
		mcp.NewTool("delete_channel_permission_overwrite",
			mcp.WithDescription("Delete a permission overwrite for a role or member on a channel"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Channel ID")),
			mcp.WithString("targetId", mcp.Required(), mcp.Description("Role ID or Member ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			targetID := getString(req.Params.Arguments, "targetId")
			if channelID == "" || targetID == "" {
				return errorResult(fmt.Errorf("channelId and targetId are required")), nil
			}

			err := client.Session.ChannelPermissionDelete(channelID, targetID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to delete permission overwrite: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Permission overwrite for `%s` deleted from channel `%s`.", targetID, channelID)), nil
		},
	)
}
