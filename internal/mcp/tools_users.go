package mcp

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/0xOpCode/discord-mcp-go/internal/discord"
	"github.com/bwmarrin/discordgo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterUserTools(s *server.MCPServer, client *discord.Client) {
	// get_user_id_by_name
	s.AddTool(
		mcp.NewTool("get_user_id_by_name",
			mcp.WithDescription("Find a Discord user ID by username or nickname in a server"),
			mcp.WithString("username", mcp.Required(), mcp.Description("Username or nickname to search")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID. Uses default if omitted.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			username := getString(req.Params.Arguments, "username")
			if username == "" {
				return errorResult(fmt.Errorf("username is required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			members, err := client.Session.GuildMembersSearch(guildID, username, 10)
			if err != nil {
				return errorResult(fmt.Errorf("failed to search members: %w", err)), nil
			}

			if len(members) == 0 {
				return successResult(fmt.Sprintf("No members found matching '%s'.", username)), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Matching Users (%d found):\n\n", len(members)))
			sb.WriteString(fmt.Sprintf("| %-20s | %-20s | %-20s |\n", "Username", "Nickname", "User ID"))
			sb.WriteString("|----------------------|----------------------|----------------------|\n")
			for _, m := range members {
				nick := m.Nick
				if nick == "" {
					nick = "N/A"
				}
				sb.WriteString(fmt.Sprintf("| %-20s | %-20s | %-20s |\n", m.User.Username, nick, m.User.ID))
			}
			return successResult(sb.String()), nil
		},
	)

	// send_private_message
	s.AddTool(
		mcp.NewTool("send_private_message",
			mcp.WithDescription("Send a direct private message (DM) to a Discord user"),
			mcp.WithString("userId", mcp.Required(), mcp.Description("Target user ID")),
			mcp.WithString("content", mcp.Required(), mcp.Description("Message content to send")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			userID := getString(req.Params.Arguments, "userId")
			content := getString(req.Params.Arguments, "content")
			if userID == "" || content == "" {
				return errorResult(fmt.Errorf("userId and content are required")), nil
			}

			dmChannel, err := client.Session.UserChannelCreate(userID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to create DM channel: %w", err)), nil
			}

			msg, err := client.Session.ChannelMessageSend(dmChannel.ID, content)
			if err != nil {
				return errorResult(fmt.Errorf("failed to send DM: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Direct message sent successfully to user `%s`. Message ID: `%s`", userID, msg.ID)), nil
		},
	)

	// edit_private_message
	s.AddTool(
		mcp.NewTool("edit_private_message",
			mcp.WithDescription("Edit a private message previously sent by the bot"),
			mcp.WithString("userId", mcp.Required(), mcp.Description("Recipient user ID")),
			mcp.WithString("messageId", mcp.Required(), mcp.Description("ID of message to edit")),
			mcp.WithString("content", mcp.Required(), mcp.Description("New message body")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			userID := getString(req.Params.Arguments, "userId")
			messageID := getString(req.Params.Arguments, "messageId")
			content := getString(req.Params.Arguments, "content")
			if userID == "" || messageID == "" || content == "" {
				return errorResult(fmt.Errorf("userId, messageId, and content are required")), nil
			}

			dmChannel, err := client.Session.UserChannelCreate(userID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to open DM channel: %w", err)), nil
			}

			msg, err := client.Session.ChannelMessageEdit(dmChannel.ID, messageID, content)
			if err != nil {
				return errorResult(fmt.Errorf("failed to edit DM: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Private message updated successfully. ID: `%s`", msg.ID)), nil
		},
	)

	// delete_private_message
	s.AddTool(
		mcp.NewTool("delete_private_message",
			mcp.WithDescription("Delete a private message previously sent by the bot"),
			mcp.WithString("userId", mcp.Required(), mcp.Description("Recipient user ID")),
			mcp.WithString("messageId", mcp.Required(), mcp.Description("ID of message to delete")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			userID := getString(req.Params.Arguments, "userId")
			messageID := getString(req.Params.Arguments, "messageId")
			if userID == "" || messageID == "" {
				return errorResult(fmt.Errorf("userId and messageId are required")), nil
			}

			dmChannel, err := client.Session.UserChannelCreate(userID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to open DM channel: %w", err)), nil
			}

			err = client.Session.ChannelMessageDelete(dmChannel.ID, messageID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to delete DM: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Private message `%s` deleted successfully.", messageID)), nil
		},
	)

	// read_private_messages
	s.AddTool(
		mcp.NewTool("read_private_messages",
			mcp.WithDescription("Read private message (DM) history with a user"),
			mcp.WithString("userId", mcp.Required(), mcp.Description("User ID to read messages from")),
			mcp.WithNumber("count", mcp.Description("Number of messages (1-100, default 50)")),
			mcp.WithString("before", mcp.Description("Cursor message ID: read before")),
			mcp.WithString("after", mcp.Description("Cursor message ID: read after")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			userID := getString(req.Params.Arguments, "userId")
			if userID == "" {
				return errorResult(fmt.Errorf("userId is required")), nil
			}

			count := getInt(req.Params.Arguments, "count", 50)
			if count < 1 {
				count = 1
			} else if count > 100 {
				count = 100
			}

			dmChannel, err := client.Session.UserChannelCreate(userID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to open DM channel: %w", err)), nil
			}

			before := getString(req.Params.Arguments, "before")
			after := getString(req.Params.Arguments, "after")

			messages, err := client.Session.ChannelMessages(dmChannel.ID, count, before, after, "")
			if err != nil {
				return errorResult(fmt.Errorf("failed to read DM messages: %w", err)), nil
			}

			if len(messages) == 0 {
				return successResult("No direct messages found."), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Direct Messages with User `%s` (%d returned):\n\n", userID, len(messages)))
			for i := len(messages) - 1; i >= 0; i-- {
				m := messages[i]
				author := "Unknown"
				if m.Author != nil {
					author = m.Author.Username
				}
				sb.WriteString(fmt.Sprintf("**[%s]** `%s` (%s):\n%s\n\n",
					m.Timestamp.Format("2006-01-02 15:04:05"),
					m.ID,
					author,
					m.Content,
				))
			}
			return successResult(sb.String()), nil
		},
	)

	// send_private_file
	s.AddTool(
		mcp.NewTool("send_private_file",
			mcp.WithDescription("Upload a file or attachment to a user's direct private message (DM)"),
			mcp.WithString("userId", mcp.Required(), mcp.Description("Target user ID")),
			mcp.WithString("filePath", mcp.Required(), mcp.Description("Local file path or base64 data URI")),
			mcp.WithString("content", mcp.Description("Optional message text accompanying file")),
			mcp.WithString("fileName", mcp.Description("Optional custom file name")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			userID := getString(req.Params.Arguments, "userId")
			filePath := getString(req.Params.Arguments, "filePath")
			if userID == "" || filePath == "" {
				return errorResult(fmt.Errorf("userId and filePath are required")), nil
			}

			dmChannel, err := client.Session.UserChannelCreate(userID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to open DM channel: %w", err)), nil
			}

			customName := getString(req.Params.Arguments, "fileName")
			fileBytes, fileName, err := readFileInput(filePath, customName)
			if err != nil {
				return errorResult(fmt.Errorf("invalid file: %w", err)), nil
			}

			content := getString(req.Params.Arguments, "content")
			msgSend := &discordgo.MessageSend{
				Content: content,
				Files: []*discordgo.File{
					{
						Name:   fileName,
						Reader: bytes.NewReader(fileBytes),
					},
				},
			}

			msg, err := client.Session.ChannelMessageSendComplex(dmChannel.ID, msgSend)
			if err != nil {
				return errorResult(fmt.Errorf("failed to send file in DM: %w", err)), nil
			}

			return successResult(fmt.Sprintf("File `%s` sent to user `%s` in DM. Message ID: `%s`", fileName, userID, msg.ID)), nil
		},
	)

	// list_guild_members
	s.AddTool(
		mcp.NewTool("list_guild_members",
			mcp.WithDescription("List members in a Discord server with roles and joined timestamps"),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
			mcp.WithNumber("limit", mcp.Description("Number of members to fetch (1-1000, default 100)")),
			mcp.WithString("after", mcp.Description("User ID cursor for pagination")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			limit := getInt(req.Params.Arguments, "limit", 100)
			if limit < 1 {
				limit = 1
			} else if limit > 1000 {
				limit = 1000
			}

			after := getString(req.Params.Arguments, "after")
			members, err := client.Session.GuildMembers(guildID, after, limit)
			if err != nil {
				return errorResult(fmt.Errorf("failed to fetch guild members: %w", err)), nil
			}

			if len(members) == 0 {
				return successResult(fmt.Sprintf("No members returned for server `%s`.", guildID)), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Server Members in `%s` (%d returned):\n\n", guildID, len(members)))
			sb.WriteString(fmt.Sprintf("| %-22s | %-20s | %-20s | %-6s |\n", "Username", "Nickname", "User ID", "Roles"))
			sb.WriteString("|------------------------|----------------------|----------------------|--------|\n")
			for _, m := range members {
				username := "Unknown"
				userID := "N/A"
				if m.User != nil {
					username = m.User.Username
					userID = m.User.ID
				}
				if len(username) > 22 {
					username = username[:19] + "..."
				}
				nick := m.Nick
				if nick == "" {
					nick = "-"
				}
				if len(nick) > 20 {
					nick = nick[:17] + "..."
				}
				sb.WriteString(fmt.Sprintf("| %-22s | %-20s | %-20s | %-6d |\n", username, nick, userID, len(m.Roles)))
			}

			return successResult(sb.String()), nil
		},
	)
}
