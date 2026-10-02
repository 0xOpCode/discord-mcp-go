package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/0xOpCode/discord-mcp-go/internal/discord"
	"github.com/bwmarrin/discordgo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterEventTools(s *server.MCPServer, client *discord.Client) {
	// create_guild_scheduled_event
	s.AddTool(
		mcp.NewTool("create_guild_scheduled_event",
			mcp.WithDescription("Create a scheduled event in a Discord server"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Event name")),
			mcp.WithString("startTime", mcp.Required(), mcp.Description("Start time (ISO-8601, e.g. 2026-10-15T18:00:00Z)")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
			mcp.WithString("description", mcp.Description("Event description")),
			mcp.WithString("channelId", mcp.Description("Voice/Stage channel ID (required if entityType is voice/stage)")),
			mcp.WithString("location", mcp.Description("External location URL/address (required if entityType is external)")),
			mcp.WithString("endTime", mcp.Description("End time (ISO-8601, required for external events)")),
			mcp.WithString("entityType", mcp.Description("Entity type: 'stage', 'voice', or 'external' (default: stage)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name := getString(req.Params.Arguments, "name")
			startTimeStr := getString(req.Params.Arguments, "startTime")
			if name == "" || startTimeStr == "" {
				return errorResult(fmt.Errorf("name and startTime are required")), nil
			}

			startTime, err := time.Parse(time.RFC3339, startTimeStr)
			if err != nil {
				return errorResult(fmt.Errorf("invalid startTime format, use RFC3339: %w", err)), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			entityTypeStr := strings.ToLower(getString(req.Params.Arguments, "entityType"))
			var entityType discordgo.GuildScheduledEventEntityType
			switch entityTypeStr {
			case "voice":
				entityType = discordgo.GuildScheduledEventEntityTypeVoice
			case "external":
				entityType = discordgo.GuildScheduledEventEntityTypeExternal
			default:
				entityType = discordgo.GuildScheduledEventEntityTypeStageInstance
			}

			params := &discordgo.GuildScheduledEventParams{
				Name:               name,
				Description:        getString(req.Params.Arguments, "description"),
				ScheduledStartTime: &startTime,
				EntityType:         entityType,
				PrivacyLevel:       discordgo.GuildScheduledEventPrivacyLevelGuildOnly,
			}

			if channelID := getString(req.Params.Arguments, "channelId"); channelID != "" {
				params.ChannelID = channelID
			}

			if location := getString(req.Params.Arguments, "location"); location != "" {
				params.EntityMetadata = &discordgo.GuildScheduledEventEntityMetadata{
					Location: location,
				}
			}

			if endTimeStr := getString(req.Params.Arguments, "endTime"); endTimeStr != "" {
				if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
					params.ScheduledEndTime = &endTime
				}
			}

			event, err := client.Session.GuildScheduledEventCreate(guildID, params)
			if err != nil {
				return errorResult(fmt.Errorf("failed to create scheduled event: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Event `%s` created successfully. ID: `%s`", event.Name, event.ID)), nil
		},
	)

	// edit_guild_scheduled_event
	s.AddTool(
		mcp.NewTool("edit_guild_scheduled_event",
			mcp.WithDescription("Modify an existing scheduled event or change its status"),
			mcp.WithString("eventId", mcp.Required(), mcp.Description("Event ID")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
			mcp.WithString("name", mcp.Description("New event name")),
			mcp.WithString("description", mcp.Description("New description")),
			mcp.WithString("status", mcp.Description("Status: 'active', 'completed', 'canceled'")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			eventID := getString(req.Params.Arguments, "eventId")
			if eventID == "" {
				return errorResult(fmt.Errorf("eventId is required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			params := &discordgo.GuildScheduledEventParams{}
			if name := getString(req.Params.Arguments, "name"); name != "" {
				params.Name = name
			}
			if desc := getString(req.Params.Arguments, "description"); desc != "" {
				params.Description = desc
			}
			if statusStr := strings.ToLower(getString(req.Params.Arguments, "status")); statusStr != "" {
				switch statusStr {
				case "active":
					params.Status = discordgo.GuildScheduledEventStatusActive
				case "completed":
					params.Status = discordgo.GuildScheduledEventStatusCompleted
				case "canceled":
					params.Status = discordgo.GuildScheduledEventStatusCanceled
				}
			}

			event, err := client.Session.GuildScheduledEventEdit(guildID, eventID, params)
			if err != nil {
				return errorResult(fmt.Errorf("failed to edit scheduled event: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Scheduled event `%s` updated successfully.", event.Name)), nil
		},
	)

	// delete_guild_scheduled_event
	s.AddTool(
		mcp.NewTool("delete_guild_scheduled_event",
			mcp.WithDescription("Delete a scheduled event from a Discord server"),
			mcp.WithString("eventId", mcp.Required(), mcp.Description("Event ID to delete")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			eventID := getString(req.Params.Arguments, "eventId")
			if eventID == "" {
				return errorResult(fmt.Errorf("eventId is required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			err = client.Session.GuildScheduledEventDelete(guildID, eventID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to delete event: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Scheduled event `%s` deleted.", eventID)), nil
		},
	)

	// list_guild_scheduled_events
	s.AddTool(
		mcp.NewTool("list_guild_scheduled_events",
			mcp.WithDescription("List all scheduled events in a Discord server"),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			events, err := client.Session.GuildScheduledEvents(guildID, true)
			if err != nil {
				return errorResult(fmt.Errorf("failed to fetch scheduled events: %w", err)), nil
			}

			if len(events) == 0 {
				return successResult("No scheduled events in this server."), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Scheduled Events in Server `%s` (%d total):\n\n", guildID, len(events)))
			sb.WriteString(fmt.Sprintf("| %-25s | %-20s | %-20s | %-8s |\n", "Event Name", "Event ID", "Start Time", "Users"))
			sb.WriteString("|---------------------------|----------------------|----------------------|----------|\n")
			for _, ev := range events {
				name := ev.Name
				if len(name) > 25 {
					name = name[:22] + "..."
				}
				sb.WriteString(fmt.Sprintf("| %-25s | %-20s | %-20s | %-8d |\n",
					name,
					ev.ID,
					ev.ScheduledStartTime.Format("2006-01-02 15:04"),
					ev.UserCount,
				))
			}
			return successResult(sb.String()), nil
		},
	)

	// get_guild_scheduled_event_users
	s.AddTool(
		mcp.NewTool("get_guild_scheduled_event_users",
			mcp.WithDescription("Get list of users interested in a scheduled event"),
			mcp.WithString("eventId", mcp.Required(), mcp.Description("Event ID")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			eventID := getString(req.Params.Arguments, "eventId")
			if eventID == "" {
				return errorResult(fmt.Errorf("eventId is required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			users, err := client.Session.GuildScheduledEventUsers(guildID, eventID, 100, false, "", "")
			if err != nil {
				return errorResult(fmt.Errorf("failed to fetch event users: %w", err)), nil
			}

			if len(users) == 0 {
				return successResult("No users subscribed to this event yet."), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Interested Users for Event `%s` (%d total):\n\n", eventID, len(users)))
			sb.WriteString(fmt.Sprintf("| %-20s | %-20s |\n", "Username", "User ID"))
			sb.WriteString("|----------------------|----------------------|\n")
			for _, u := range users {
				if u.User != nil {
					sb.WriteString(fmt.Sprintf("| %-20s | %-20s |\n", u.User.Username, u.User.ID))
				}
			}
			return successResult(sb.String()), nil
		},
	)
}
