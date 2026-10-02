package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/0xOpCode/discord-mcp-go/internal/discord"
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
}
