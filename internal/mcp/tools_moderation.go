package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/0xOpCode/discord-mcp-go/internal/discord"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterModerationTools(s *server.MCPServer, client *discord.Client) {
	// kick_member
	s.AddTool(
		mcp.NewTool("kick_member",
			mcp.WithDescription("Kick a member from a Discord server"),
			mcp.WithString("userId", mcp.Required(), mcp.Description("User ID to kick")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
			mcp.WithString("reason", mcp.Description("Audit log reason for kick")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			userID := getString(req.Params.Arguments, "userId")
			if userID == "" {
				return errorResult(fmt.Errorf("userId is required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			reason := getString(req.Params.Arguments, "reason")

			err = client.Session.GuildMemberDeleteWithReason(guildID, userID, reason)
			if err != nil {
				return errorResult(fmt.Errorf("failed to kick user: %w", err)), nil
			}

			return successResult(fmt.Sprintf("User `%s` kicked from server `%s`.", userID, guildID)), nil
		},
	)

	// ban_member
	s.AddTool(
		mcp.NewTool("ban_member",
			mcp.WithDescription("Ban a user from a Discord server"),
			mcp.WithString("userId", mcp.Required(), mcp.Description("User ID to ban")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
			mcp.WithNumber("deleteMessageDays", mcp.Description("Number of days of messages to delete (0-7)")),
			mcp.WithString("reason", mcp.Description("Audit log reason for ban")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			userID := getString(req.Params.Arguments, "userId")
			if userID == "" {
				return errorResult(fmt.Errorf("userId is required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			days := getInt(req.Params.Arguments, "deleteMessageDays", 0)
			reason := getString(req.Params.Arguments, "reason")

			err = client.Session.GuildBanCreateWithReason(guildID, userID, reason, days)
			if err != nil {
				return errorResult(fmt.Errorf("failed to ban user: %w", err)), nil
			}

			return successResult(fmt.Sprintf("User `%s` banned from server `%s`.", userID, guildID)), nil
		},
	)

	// unban_member
	s.AddTool(
		mcp.NewTool("unban_member",
			mcp.WithDescription("Remove a ban from a user in a Discord server"),
			mcp.WithString("userId", mcp.Required(), mcp.Description("User ID to unban")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			userID := getString(req.Params.Arguments, "userId")
			if userID == "" {
				return errorResult(fmt.Errorf("userId is required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			err = client.Session.GuildBanDelete(guildID, userID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to unban user: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Ban removed for user `%s` in server `%s`.", userID, guildID)), nil
		},
	)

	// timeout_member
	s.AddTool(
		mcp.NewTool("timeout_member",
			mcp.WithDescription("Timeout (mute) a member in a Discord server"),
			mcp.WithString("userId", mcp.Required(), mcp.Description("User ID")),
			mcp.WithNumber("durationMinutes", mcp.Required(), mcp.Description("Timeout duration in minutes")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			userID := getString(req.Params.Arguments, "userId")
			duration := getInt(req.Params.Arguments, "durationMinutes", 0)
			if userID == "" || duration <= 0 {
				return errorResult(fmt.Errorf("userId and positive durationMinutes are required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			until := time.Now().Add(time.Duration(duration) * time.Minute)

			err = client.Session.GuildMemberTimeout(guildID, userID, &until)
			if err != nil {
				return errorResult(fmt.Errorf("failed to timeout member: %w", err)), nil
			}

			return successResult(fmt.Sprintf("User `%s` timed out for %d minutes (until %s).", userID, duration, until.Format(time.RFC3339))), nil
		},
	)

	// remove_timeout
	s.AddTool(
		mcp.NewTool("remove_timeout",
			mcp.WithDescription("Remove timeout (unmute) from a server member"),
			mcp.WithString("userId", mcp.Required(), mcp.Description("User ID")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			userID := getString(req.Params.Arguments, "userId")
			if userID == "" {
				return errorResult(fmt.Errorf("userId is required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			err = client.Session.GuildMemberTimeout(guildID, userID, nil)
			if err != nil {
				return errorResult(fmt.Errorf("failed to remove timeout: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Timeout removed for user `%s` in server `%s`.", userID, guildID)), nil
		},
	)

	// set_nickname
	s.AddTool(
		mcp.NewTool("set_nickname",
			mcp.WithDescription("Set or clear a server member's nickname"),
			mcp.WithString("userId", mcp.Required(), mcp.Description("User ID")),
			mcp.WithString("nickname", mcp.Description("New nickname (empty to reset to username)")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			userID := getString(req.Params.Arguments, "userId")
			if userID == "" {
				return errorResult(fmt.Errorf("userId is required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			nickname := getString(req.Params.Arguments, "nickname")

			err = client.Session.GuildMemberNickname(guildID, userID, nickname)
			if err != nil {
				return errorResult(fmt.Errorf("failed to set nickname: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Nickname updated for user `%s`.", userID)), nil
		},
	)

	// get_bans
	s.AddTool(
		mcp.NewTool("get_bans",
			mcp.WithDescription("Get list of banned users in a Discord server"),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			bans, err := client.Session.GuildBans(guildID, 100, "", "")
			if err != nil {
				return errorResult(fmt.Errorf("failed to fetch bans: %w", err)), nil
			}

			if len(bans) == 0 {
				return successResult("No banned users in this server."), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Banned Users in Server `%s` (%d total):\n\n", guildID, len(bans)))
			sb.WriteString(fmt.Sprintf("| %-20s | %-20s | %-30s |\n", "Username", "User ID", "Ban Reason"))
			sb.WriteString("|----------------------|----------------------|--------------------------------|\n")
			for _, b := range bans {
				username := "Unknown"
				userID := "Unknown"
				if b.User != nil {
					username = b.User.Username
					userID = b.User.ID
				}
				reason := b.Reason
				if reason == "" {
					reason = "None provided"
				}
				sb.WriteString(fmt.Sprintf("| %-20s | %-20s | %-30s |\n", username, userID, reason))
			}
			return successResult(sb.String()), nil
		},
	)
}
