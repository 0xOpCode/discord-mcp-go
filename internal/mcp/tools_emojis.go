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

func RegisterEmojiTools(s *server.MCPServer, client *discord.Client) {
	// list_emojis
	s.AddTool(
		mcp.NewTool("list_emojis",
			mcp.WithDescription("List all custom emojis in a Discord server"),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			emojis, err := client.Session.GuildEmojis(guildID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to fetch emojis: %w", err)), nil
			}

			if len(emojis) == 0 {
				return successResult("No custom emojis found in this server."), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Custom Emojis in Server `%s` (%d total):\n\n", guildID, len(emojis)))
			sb.WriteString(fmt.Sprintf("| %-20s | %-20s | %-10s |\n", "Name", "Emoji ID", "Animated"))
			sb.WriteString("|----------------------|----------------------|------------|\n")
			for _, e := range emojis {
				sb.WriteString(fmt.Sprintf("| %-20s | %-20s | %-10t |\n", e.Name, e.ID, e.Animated))
			}
			return successResult(sb.String()), nil
		},
	)

	// get_emoji_details
	s.AddTool(
		mcp.NewTool("get_emoji_details",
			mcp.WithDescription("Get detailed information about a custom emoji"),
			mcp.WithString("emojiId", mcp.Required(), mcp.Description("Emoji ID")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			emojiID := getString(req.Params.Arguments, "emojiId")
			if emojiID == "" {
				return errorResult(fmt.Errorf("emojiId is required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			emoji, err := client.Session.GuildEmoji(guildID, emojiID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to fetch emoji: %w", err)), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Emoji: :%s:\n", emoji.Name))
			sb.WriteString(fmt.Sprintf("- **ID**: `%s`\n", emoji.ID))
			sb.WriteString(fmt.Sprintf("- **Animated**: %t\n", emoji.Animated))
			sb.WriteString(fmt.Sprintf("- **Managed**: %t\n", emoji.Managed))
			sb.WriteString(fmt.Sprintf("- **Require Colons**: %t\n", emoji.RequireColons))
			if emoji.User != nil {
				sb.WriteString(fmt.Sprintf("- **Uploaded By**: %s (`%s`)\n", emoji.User.Username, emoji.User.ID))
			}
			return successResult(sb.String()), nil
		},
	)

	// create_emoji
	s.AddTool(
		mcp.NewTool("create_emoji",
			mcp.WithDescription("Upload a new custom emoji to a Discord server"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Emoji name")),
			mcp.WithString("image", mcp.Required(), mcp.Description("Base64 data URI image string (max 256KB)")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name := getString(req.Params.Arguments, "name")
			image := getString(req.Params.Arguments, "image")
			if name == "" || image == "" {
				return errorResult(fmt.Errorf("name and image are required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			params := &discordgo.EmojiParams{
				Name:  name,
				Image: image,
			}

			emoji, err := client.Session.GuildEmojiCreate(guildID, params)
			if err != nil {
				return errorResult(fmt.Errorf("failed to create emoji: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Emoji `:%s:` created. ID: `%s`", emoji.Name, emoji.ID)), nil
		},
	)

	// edit_emoji
	s.AddTool(
		mcp.NewTool("edit_emoji",
			mcp.WithDescription("Rename a custom emoji on a server"),
			mcp.WithString("emojiId", mcp.Required(), mcp.Description("Emoji ID")),
			mcp.WithString("name", mcp.Required(), mcp.Description("New emoji name")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			emojiID := getString(req.Params.Arguments, "emojiId")
			name := getString(req.Params.Arguments, "name")
			if emojiID == "" || name == "" {
				return errorResult(fmt.Errorf("emojiId and name are required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			params := &discordgo.EmojiParams{
				Name: name,
			}

			emoji, err := client.Session.GuildEmojiEdit(guildID, emojiID, params)
			if err != nil {
				return errorResult(fmt.Errorf("failed to edit emoji: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Emoji updated: `:%s:`", emoji.Name)), nil
		},
	)

	// delete_emoji
	s.AddTool(
		mcp.NewTool("delete_emoji",
			mcp.WithDescription("Delete a custom emoji from a server"),
			mcp.WithString("emojiId", mcp.Required(), mcp.Description("Emoji ID to delete")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			emojiID := getString(req.Params.Arguments, "emojiId")
			if emojiID == "" {
				return errorResult(fmt.Errorf("emojiId is required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			err = client.Session.GuildEmojiDelete(guildID, emojiID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to delete emoji: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Emoji `%s` deleted from server `%s`.", emojiID, guildID)), nil
		},
	)
}
