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

func RegisterChannelTools(s *server.MCPServer, client *discord.Client) {
	// create_text_channel
	s.AddTool(
		mcp.NewTool("create_text_channel",
			mcp.WithDescription("Create a new text channel in a Discord server"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Channel name")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
			mcp.WithString("topic", mcp.Description("Channel topic")),
			mcp.WithString("categoryId", mcp.Description("Parent category ID")),
			mcp.WithBoolean("nsfw", mcp.Description("Whether the channel is NSFW")),
			mcp.WithNumber("slowmode", mcp.Description("Slowmode delay in seconds (0-21600)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name := getString(req.Params.Arguments, "name")
			if name == "" {
				return errorResult(fmt.Errorf("channel name is required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			topic := getString(req.Params.Arguments, "topic")
			categoryID := getString(req.Params.Arguments, "categoryId")
			nsfw := getBool(req.Params.Arguments, "nsfw", false)
			slowmode := getInt(req.Params.Arguments, "slowmode", 0)

			data := discordgo.GuildChannelCreateData{
				Name:             name,
				Type:             discordgo.ChannelTypeGuildText,
				Topic:            topic,
				ParentID:         categoryID,
				NSFW:             nsfw,
				RateLimitPerUser: slowmode,
			}

			ch, err := client.Session.GuildChannelCreateComplex(guildID, data)
			if err != nil {
				return errorResult(fmt.Errorf("failed to create channel: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Text channel `%s` created successfully. ID: `%s`", ch.Name, ch.ID)), nil
		},
	)

	// edit_text_channel
	s.AddTool(
		mcp.NewTool("edit_text_channel",
			mcp.WithDescription("Edit settings of a text channel"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Channel ID to edit")),
			mcp.WithString("name", mcp.Description("New channel name")),
			mcp.WithString("topic", mcp.Description("New channel topic")),
			mcp.WithBoolean("nsfw", mcp.Description("NSFW flag")),
			mcp.WithNumber("slowmode", mcp.Description("Slowmode delay in seconds")),
			mcp.WithString("categoryId", mcp.Description("Parent category ID")),
			mcp.WithNumber("position", mcp.Description("Channel position")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			if channelID == "" {
				return errorResult(fmt.Errorf("channelId is required")), nil
			}

			edit := &discordgo.ChannelEdit{}
			if name := getString(req.Params.Arguments, "name"); name != "" {
				edit.Name = name
			}
			if topic := getString(req.Params.Arguments, "topic"); topic != "" {
				edit.Topic = topic
			}
			if catID := getString(req.Params.Arguments, "categoryId"); catID != "" {
				edit.ParentID = catID
			}
			if _, ok := req.Params.Arguments["nsfw"]; ok {
				nsfw := getBool(req.Params.Arguments, "nsfw", false)
				edit.NSFW = &nsfw
			}
			if _, ok := req.Params.Arguments["slowmode"]; ok {
				slowmode := getInt(req.Params.Arguments, "slowmode", 0)
				edit.RateLimitPerUser = &slowmode
			}
			if _, ok := req.Params.Arguments["position"]; ok {
				pos := getInt(req.Params.Arguments, "position", 0)
				edit.Position = &pos
			}

			ch, err := client.Session.ChannelEditComplex(channelID, edit)
			if err != nil {
				return errorResult(fmt.Errorf("failed to edit channel: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Channel `%s` updated successfully. ID: `%s`", ch.Name, ch.ID)), nil
		},
	)

	// delete_channel
	s.AddTool(
		mcp.NewTool("delete_channel",
			mcp.WithDescription("Delete a channel from a Discord server"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Channel ID to delete")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			if channelID == "" {
				return errorResult(fmt.Errorf("channelId is required")), nil
			}

			ch, err := client.Session.ChannelDelete(channelID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to delete channel: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Channel `%s` (`%s`) deleted successfully.", ch.Name, ch.ID)), nil
		},
	)

	// find_channel
	s.AddTool(
		mcp.NewTool("find_channel",
			mcp.WithDescription("Find channel ID and details by name in a Discord server"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Channel name to search")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name := strings.ToLower(getString(req.Params.Arguments, "name"))
			if name == "" {
				return errorResult(fmt.Errorf("channel name is required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			channels, err := client.Session.GuildChannels(guildID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to fetch channels: %w", err)), nil
			}

			var matches []*discordgo.Channel
			for _, ch := range channels {
				if strings.Contains(strings.ToLower(ch.Name), name) {
					matches = append(matches, ch)
				}
			}

			if len(matches) == 0 {
				return successResult(fmt.Sprintf("No channels found matching '%s'.", name)), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Channels Matching '%s' (%d found):\n\n", name, len(matches)))
			sb.WriteString(fmt.Sprintf("| %-25s | %-20s | %-12s |\n", "Channel Name", "Channel ID", "Type"))
			sb.WriteString("|---------------------------|----------------------|--------------|\n")
			for _, ch := range matches {
				sb.WriteString(fmt.Sprintf("| %-25s | %-20s | %-12s |\n", ch.Name, ch.ID, channelTypeToString(ch.Type)))
			}
			return successResult(sb.String()), nil
		},
	)

	// list_channels
	s.AddTool(
		mcp.NewTool("list_channels",
			mcp.WithDescription("List all channels in a Discord server"),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			channels, err := client.Session.GuildChannels(guildID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to fetch channels: %w", err)), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Server Channels (%d total):\n\n", len(channels)))
			sb.WriteString(fmt.Sprintf("| %-25s | %-20s | %-12s | %-10s |\n", "Name", "ID", "Type", "Position"))
			sb.WriteString("|---------------------------|----------------------|--------------|------------|\n")
			for _, ch := range channels {
				name := ch.Name
				if len(name) > 25 {
					name = name[:22] + "..."
				}
				sb.WriteString(fmt.Sprintf("| %-25s | %-20s | %-12s | %-10d |\n", name, ch.ID, channelTypeToString(ch.Type), ch.Position))
			}
			return successResult(sb.String()), nil
		},
	)

	// get_channel_info
	s.AddTool(
		mcp.NewTool("get_channel_info",
			mcp.WithDescription("Get detailed information about a Discord channel"),
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

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Channel: #%s\n", ch.Name))
			sb.WriteString(fmt.Sprintf("- **ID**: `%s`\n", ch.ID))
			sb.WriteString(fmt.Sprintf("- **Type**: %s\n", channelTypeToString(ch.Type)))
			sb.WriteString(fmt.Sprintf("- **Guild ID**: `%s`\n", ch.GuildID))
			sb.WriteString(fmt.Sprintf("- **Position**: %d\n", ch.Position))
			if ch.ParentID != "" {
				sb.WriteString(fmt.Sprintf("- **Category ID**: `%s`\n", ch.ParentID))
			}
			if ch.Topic != "" {
				sb.WriteString(fmt.Sprintf("- **Topic**: %s\n", ch.Topic))
			}
			sb.WriteString(fmt.Sprintf("- **NSFW**: %t\n", ch.NSFW))
			sb.WriteString(fmt.Sprintf("- **Slowmode Delay**: %d seconds\n", ch.RateLimitPerUser))
			return successResult(sb.String()), nil
		},
	)

	// move_channel
	s.AddTool(
		mcp.NewTool("move_channel",
			mcp.WithDescription("Move a channel to another category or change its position"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Channel ID")),
			mcp.WithString("categoryId", mcp.Description("New Category ID")),
			mcp.WithNumber("position", mcp.Description("New position index")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			if channelID == "" {
				return errorResult(fmt.Errorf("channelId is required")), nil
			}

			edit := &discordgo.ChannelEdit{}
			if catID := getString(req.Params.Arguments, "categoryId"); catID != "" {
				edit.ParentID = catID
			}
			if _, ok := req.Params.Arguments["position"]; ok {
				pos := getInt(req.Params.Arguments, "position", 0)
				edit.Position = &pos
			}

			ch, err := client.Session.ChannelEditComplex(channelID, edit)
			if err != nil {
				return errorResult(fmt.Errorf("failed to move channel: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Channel `%s` moved successfully.", ch.Name)), nil
		},
	)
}
