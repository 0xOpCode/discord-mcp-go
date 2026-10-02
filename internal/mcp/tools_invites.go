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

func RegisterInviteTools(s *server.MCPServer, client *discord.Client) {
	// create_invite
	s.AddTool(
		mcp.NewTool("create_invite",
			mcp.WithDescription("Create an invite link for a Discord channel"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Channel ID")),
			mcp.WithNumber("maxAge", mcp.Description("Duration of invite in seconds (0 for permanent, default 86400)")),
			mcp.WithNumber("maxUses", mcp.Description("Maximum number of uses (0 for unlimited)")),
			mcp.WithBoolean("temporary", mcp.Description("Grant temporary membership")),
			mcp.WithBoolean("unique", mcp.Description("Guarantee unique invite code")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			if channelID == "" {
				return errorResult(fmt.Errorf("channelId is required")), nil
			}

			maxAge := getInt(req.Params.Arguments, "maxAge", 86400)
			maxUses := getInt(req.Params.Arguments, "maxUses", 0)
			temporary := getBool(req.Params.Arguments, "temporary", false)
			unique := getBool(req.Params.Arguments, "unique", false)

			inviteData := discordgo.Invite{
				MaxAge:    maxAge,
				MaxUses:   maxUses,
				Temporary: temporary,
				Unique:    unique,
			}

			invite, err := client.Session.ChannelInviteCreate(channelID, inviteData)
			if err != nil {
				return errorResult(fmt.Errorf("failed to create invite: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Invite created: https://discord.gg/%s (Code: `%s`)", invite.Code, invite.Code)), nil
		},
	)

	// list_invites
	s.AddTool(
		mcp.NewTool("list_invites",
			mcp.WithDescription("List all active invites on a Discord server"),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			invites, err := client.Session.GuildInvites(guildID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to fetch invites: %w", err)), nil
			}

			if len(invites) == 0 {
				return successResult("No active invites found on this server."), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Active Server Invites (%d total):\n\n", len(invites)))
			sb.WriteString(fmt.Sprintf("| %-12s | %-20s | %-6s | %-8s |\n", "Invite Code", "Channel ID", "Uses", "Max Uses"))
			sb.WriteString("|--------------|----------------------|--------|----------|\n")
			for _, inv := range invites {
				chID := "N/A"
				if inv.Channel != nil {
					chID = inv.Channel.ID
				}
				sb.WriteString(fmt.Sprintf("| %-12s | %-20s | %-6d | %-8d |\n", inv.Code, chID, inv.Uses, inv.MaxUses))
			}
			return successResult(sb.String()), nil
		},
	)

	// delete_invite
	s.AddTool(
		mcp.NewTool("delete_invite",
			mcp.WithDescription("Delete (revoke) an invite code so it can no longer be used"),
			mcp.WithString("code", mcp.Required(), mcp.Description("Invite code to delete")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			code := getString(req.Params.Arguments, "code")
			if code == "" {
				return errorResult(fmt.Errorf("invite code is required")), nil
			}

			// Strip full url prefix if user passed URL
			code = strings.TrimPrefix(code, "https://discord.gg/")
			code = strings.TrimPrefix(code, "discord.gg/")

			_, err := client.Session.InviteDelete(code)
			if err != nil {
				return errorResult(fmt.Errorf("failed to delete invite: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Invite `%s` deleted successfully.", code)), nil
		},
	)

	// get_invite_details
	s.AddTool(
		mcp.NewTool("get_invite_details",
			mcp.WithDescription("Get information about a Discord invite link or code"),
			mcp.WithString("code", mcp.Required(), mcp.Description("Invite code or full URL")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			code := getString(req.Params.Arguments, "code")
			if code == "" {
				return errorResult(fmt.Errorf("invite code is required")), nil
			}

			code = strings.TrimPrefix(code, "https://discord.gg/")
			code = strings.TrimPrefix(code, "discord.gg/")

			invite, err := client.Session.InviteWithCounts(code)
			if err != nil {
				return errorResult(fmt.Errorf("failed to fetch invite details: %w", err)), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Invite Details: `%s`\n", invite.Code))
			if invite.Guild != nil {
				sb.WriteString(fmt.Sprintf("- **Guild**: %s (`%s`)\n", invite.Guild.Name, invite.Guild.ID))
			}
			if invite.Channel != nil {
				sb.WriteString(fmt.Sprintf("- **Channel**: #%s (`%s`)\n", invite.Channel.Name, invite.Channel.ID))
			}
			sb.WriteString(fmt.Sprintf("- **Approximate Members**: %d\n", invite.ApproximateMemberCount))
			sb.WriteString(fmt.Sprintf("- **Approximate Online**: %d\n", invite.ApproximatePresenceCount))
			return successResult(sb.String()), nil
		},
	)
}
