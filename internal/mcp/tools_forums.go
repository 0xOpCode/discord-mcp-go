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

func RegisterForumTools(s *server.MCPServer, client *discord.Client) {
	// create_forum_channel
	s.AddTool(
		mcp.NewTool("create_forum_channel",
			mcp.WithDescription("Create a new forum channel in a Discord server"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Forum channel name")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
			mcp.WithString("topic", mcp.Description("Forum channel topic/guidelines")),
			mcp.WithString("categoryId", mcp.Description("Parent category ID")),
			mcp.WithBoolean("nsfw", mcp.Description("NSFW flag")),
			mcp.WithNumber("slowmode", mcp.Description("Slowmode delay in seconds")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name := getString(req.Params.Arguments, "name")
			if name == "" {
				return errorResult(fmt.Errorf("name is required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			data := discordgo.GuildChannelCreateData{
				Name:             name,
				Type:             discordgo.ChannelTypeGuildForum,
				Topic:            getString(req.Params.Arguments, "topic"),
				ParentID:         getString(req.Params.Arguments, "categoryId"),
				NSFW:             getBool(req.Params.Arguments, "nsfw", false),
				RateLimitPerUser: getInt(req.Params.Arguments, "slowmode", 0),
			}

			ch, err := client.Session.GuildChannelCreateComplex(guildID, data)
			if err != nil {
				return errorResult(fmt.Errorf("failed to create forum channel: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Forum channel `%s` created successfully. ID: `%s`", ch.Name, ch.ID)), nil
		},
	)

	// edit_forum_channel
	s.AddTool(
		mcp.NewTool("edit_forum_channel",
			mcp.WithDescription("Edit settings of a forum channel"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Forum channel ID")),
			mcp.WithString("name", mcp.Description("New channel name")),
			mcp.WithString("topic", mcp.Description("New topic guidelines")),
			mcp.WithBoolean("nsfw", mcp.Description("NSFW flag")),
			mcp.WithNumber("slowmode", mcp.Description("Slowmode delay in seconds")),
			mcp.WithString("categoryId", mcp.Description("Parent category ID")),
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
				sm := getInt(req.Params.Arguments, "slowmode", 0)
				edit.RateLimitPerUser = &sm
			}

			ch, err := client.Session.ChannelEditComplex(channelID, edit)
			if err != nil {
				return errorResult(fmt.Errorf("failed to edit forum channel: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Forum channel `%s` updated successfully.", ch.Name)), nil
		},
	)

	// list_forum_channels
	s.AddTool(
		mcp.NewTool("list_forum_channels",
			mcp.WithDescription("List all forum channels in a Discord server"),
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

			var forums []*discordgo.Channel
			for _, ch := range channels {
				if ch.Type == discordgo.ChannelTypeGuildForum {
					forums = append(forums, ch)
				}
			}

			if len(forums) == 0 {
				return successResult("No forum channels found in this server."), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Forum Channels in Server `%s` (%d total):\n\n", guildID, len(forums)))
			sb.WriteString(fmt.Sprintf("| %-25s | %-20s | %-6s |\n", "Forum Name", "Channel ID", "Tags"))
			sb.WriteString("|---------------------------|----------------------|--------|\n")
			for _, f := range forums {
				sb.WriteString(fmt.Sprintf("| %-25s | %-20s | %-6d |\n", f.Name, f.ID, len(f.AvailableTags)))
			}
			return successResult(sb.String()), nil
		},
	)

	// get_forum_channel_info
	s.AddTool(
		mcp.NewTool("get_forum_channel_info",
			mcp.WithDescription("Get detailed information about a forum channel including tags"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Forum channel ID")),
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
			sb.WriteString(fmt.Sprintf("### Forum Channel: %s\n", ch.Name))
			sb.WriteString(fmt.Sprintf("- **ID**: `%s`\n", ch.ID))
			sb.WriteString(fmt.Sprintf("- **Guild ID**: `%s`\n", ch.GuildID))
			sb.WriteString(fmt.Sprintf("- **Topic**: %s\n", ch.Topic))
			sb.WriteString(fmt.Sprintf("- **Available Tags (%d)**:\n", len(ch.AvailableTags)))
			for _, tag := range ch.AvailableTags {
				sb.WriteString(fmt.Sprintf("  - `%s`: %s (Moderated: %t)\n", tag.ID, tag.Name, tag.Moderated))
			}
			return successResult(sb.String()), nil
		},
	)

	// list_forum_tags
	s.AddTool(
		mcp.NewTool("list_forum_tags",
			mcp.WithDescription("List all available tags configured on a forum channel"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Forum channel ID")),
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

			if len(ch.AvailableTags) == 0 {
				return successResult(fmt.Sprintf("No tags configured on forum channel `%s`.", ch.Name)), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Tags on Forum Channel `%s` (%d total):\n\n", ch.Name, len(ch.AvailableTags)))
			sb.WriteString(fmt.Sprintf("| %-20s | %-20s | %-10s |\n", "Tag Name", "Tag ID", "Moderated"))
			sb.WriteString("|----------------------|----------------------|------------|\n")
			for _, t := range ch.AvailableTags {
				sb.WriteString(fmt.Sprintf("| %-20s | %-20s | %-10t |\n", t.Name, t.ID, t.Moderated))
			}
			return successResult(sb.String()), nil
		},
	)

	// create_forum_post
	s.AddTool(
		mcp.NewTool("create_forum_post",
			mcp.WithDescription("Create a new thread post with starter message in a forum channel"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Forum channel ID")),
			mcp.WithString("title", mcp.Required(), mcp.Description("Post thread title")),
			mcp.WithString("content", mcp.Required(), mcp.Description("Initial message content")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			title := getString(req.Params.Arguments, "title")
			content := getString(req.Params.Arguments, "content")
			if channelID == "" || title == "" || content == "" {
				return errorResult(fmt.Errorf("channelId, title, and content are required")), nil
			}

			thread, err := client.Session.ForumThreadStart(channelID, title, 1440, content)
			if err != nil {
				return errorResult(fmt.Errorf("failed to create forum post: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Forum post `%s` created. Thread ID: `%s`", thread.Name, thread.ID)), nil
		},
	)

	// list_forum_posts
	s.AddTool(
		mcp.NewTool("list_forum_posts",
			mcp.WithDescription("List active posts (threads) in a forum channel"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Forum channel ID")),
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

			threadsList, err := client.Session.GuildThreadsActive(ch.GuildID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to fetch active threads: %w", err)), nil
			}

			var matches []*discordgo.Channel
			for _, th := range threadsList.Threads {
				if th.ParentID == channelID {
					matches = append(matches, th)
				}
			}

			if len(matches) == 0 {
				return successResult(fmt.Sprintf("No active posts found in forum channel `%s`.", ch.Name)), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Active Posts in Forum `%s` (%d total):\n\n", ch.Name, len(matches)))
			sb.WriteString(fmt.Sprintf("| %-30s | %-20s | %-8s |\n", "Post Title", "Thread ID", "Messages"))
			sb.WriteString("|--------------------------------|----------------------|----------|\n")
			for _, th := range matches {
				name := th.Name
				if len(name) > 30 {
					name = name[:27] + "..."
				}
				sb.WriteString(fmt.Sprintf("| %-30s | %-20s | %-8d |\n", name, th.ID, th.MessageCount))
			}
			return successResult(sb.String()), nil
		},
	)

	// modify_forum_post
	s.AddTool(
		mcp.NewTool("modify_forum_post",
			mcp.WithDescription("Modify a forum post thread (lock, archive, rename)"),
			mcp.WithString("threadId", mcp.Required(), mcp.Description("Post thread ID")),
			mcp.WithString("name", mcp.Description("New post title")),
			mcp.WithBoolean("locked", mcp.Description("Lock/unlock the thread")),
			mcp.WithBoolean("archived", mcp.Description("Archive/unarchive the thread")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			threadID := getString(req.Params.Arguments, "threadId")
			if threadID == "" {
				return errorResult(fmt.Errorf("threadId is required")), nil
			}

			edit := &discordgo.ChannelEdit{}
			if name := getString(req.Params.Arguments, "name"); name != "" {
				edit.Name = name
			}
			if _, ok := req.Params.Arguments["locked"]; ok {
				locked := getBool(req.Params.Arguments, "locked", false)
				edit.Locked = &locked
			}
			if _, ok := req.Params.Arguments["archived"]; ok {
				archived := getBool(req.Params.Arguments, "archived", false)
				edit.Archived = &archived
			}

			th, err := client.Session.ChannelEditComplex(threadID, edit)
			if err != nil {
				return errorResult(fmt.Errorf("failed to modify forum thread: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Forum post `%s` updated successfully.", th.Name)), nil
		},
	)
}
