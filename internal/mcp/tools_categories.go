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

func RegisterCategoryTools(s *server.MCPServer, client *discord.Client) {
	// create_category
	s.AddTool(
		mcp.NewTool("create_category",
			mcp.WithDescription("Create a new channel category in a Discord server"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Category name")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
			mcp.WithNumber("position", mcp.Description("Category position index")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name := getString(req.Params.Arguments, "name")
			if name == "" {
				return errorResult(fmt.Errorf("category name is required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			position := getInt(req.Params.Arguments, "position", 0)

			data := discordgo.GuildChannelCreateData{
				Name:     name,
				Type:     discordgo.ChannelTypeGuildCategory,
				Position: position,
			}

			cat, err := client.Session.GuildChannelCreateComplex(guildID, data)
			if err != nil {
				return errorResult(fmt.Errorf("failed to create category: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Category `%s` created successfully. ID: `%s`", cat.Name, cat.ID)), nil
		},
	)

	// edit_category
	s.AddTool(
		mcp.NewTool("edit_category",
			mcp.WithDescription("Edit an existing channel category"),
			mcp.WithString("categoryId", mcp.Required(), mcp.Description("Category ID")),
			mcp.WithString("name", mcp.Description("New category name")),
			mcp.WithNumber("position", mcp.Description("New position index")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			categoryID := getString(req.Params.Arguments, "categoryId")
			if categoryID == "" {
				return errorResult(fmt.Errorf("categoryId is required")), nil
			}

			edit := &discordgo.ChannelEdit{}
			if name := getString(req.Params.Arguments, "name"); name != "" {
				edit.Name = name
			}
			if _, ok := req.Params.Arguments["position"]; ok {
				pos := getInt(req.Params.Arguments, "position", 0)
				edit.Position = &pos
			}

			cat, err := client.Session.ChannelEditComplex(categoryID, edit)
			if err != nil {
				return errorResult(fmt.Errorf("failed to edit category: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Category `%s` updated successfully. ID: `%s`", cat.Name, cat.ID)), nil
		},
	)

	// delete_category
	s.AddTool(
		mcp.NewTool("delete_category",
			mcp.WithDescription("Delete a channel category"),
			mcp.WithString("categoryId", mcp.Required(), mcp.Description("Category ID to delete")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			categoryID := getString(req.Params.Arguments, "categoryId")
			if categoryID == "" {
				return errorResult(fmt.Errorf("categoryId is required")), nil
			}

			cat, err := client.Session.ChannelDelete(categoryID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to delete category: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Category `%s` (`%s`) deleted successfully.", cat.Name, cat.ID)), nil
		},
	)

	// find_category
	s.AddTool(
		mcp.NewTool("find_category",
			mcp.WithDescription("Find category ID by name in a Discord server"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Category name to find")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name := strings.ToLower(getString(req.Params.Arguments, "name"))
			if name == "" {
				return errorResult(fmt.Errorf("category name is required")), nil
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
				if ch.Type == discordgo.ChannelTypeGuildCategory && strings.Contains(strings.ToLower(ch.Name), name) {
					matches = append(matches, ch)
				}
			}

			if len(matches) == 0 {
				return successResult(fmt.Sprintf("No categories found matching '%s'.", name)), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Categories Matching '%s' (%d found):\n\n", name, len(matches)))
			sb.WriteString(fmt.Sprintf("| %-25s | %-20s |\n", "Category Name", "Category ID"))
			sb.WriteString("|---------------------------|----------------------|\n")
			for _, cat := range matches {
				sb.WriteString(fmt.Sprintf("| %-25s | %-20s |\n", cat.Name, cat.ID))
			}
			return successResult(sb.String()), nil
		},
	)

	// list_channels_in_category
	s.AddTool(
		mcp.NewTool("list_channels_in_category",
			mcp.WithDescription("List all channels belonging to a specific category"),
			mcp.WithString("categoryId", mcp.Required(), mcp.Description("Category ID")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			categoryID := getString(req.Params.Arguments, "categoryId")
			if categoryID == "" {
				return errorResult(fmt.Errorf("categoryId is required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			channels, err := client.Session.GuildChannels(guildID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to fetch channels: %w", err)), nil
			}

			var children []*discordgo.Channel
			for _, ch := range channels {
				if ch.ParentID == categoryID {
					children = append(children, ch)
				}
			}

			if len(children) == 0 {
				return successResult(fmt.Sprintf("No channels found inside category `%s`.", categoryID)), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Channels in Category `%s` (%d found):\n\n", categoryID, len(children)))
			sb.WriteString(fmt.Sprintf("| %-25s | %-20s | %-12s |\n", "Channel Name", "Channel ID", "Type"))
			sb.WriteString("|---------------------------|----------------------|--------------|\n")
			for _, ch := range children {
				sb.WriteString(fmt.Sprintf("| %-25s | %-20s | %-12s |\n", ch.Name, ch.ID, channelTypeToString(ch.Type)))
			}
			return successResult(sb.String()), nil
		},
	)
}
