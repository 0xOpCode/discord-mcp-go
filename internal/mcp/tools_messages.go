package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/0xOpCode/discord-mcp-go/internal/discord"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterMessageTools(s *server.MCPServer, client *discord.Client) {
	// send_message
	s.AddTool(
		mcp.NewTool("send_message",
			mcp.WithDescription("Send a message to a specific Discord text channel"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Channel ID where message will be posted")),
			mcp.WithString("content", mcp.Required(), mcp.Description("Text message body")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			content := getString(req.Params.Arguments, "content")
			if channelID == "" || content == "" {
				return errorResult(fmt.Errorf("channelId and content are required")), nil
			}

			msg, err := client.Session.ChannelMessageSend(channelID, content)
			if err != nil {
				return errorResult(fmt.Errorf("failed to send message: %w", err)), nil
			}
			return successResult(fmt.Sprintf("Message sent successfully. ID: `%s`", msg.ID)), nil
		},
	)

	// edit_message
	s.AddTool(
		mcp.NewTool("edit_message",
			mcp.WithDescription("Edit an existing message previously sent by the bot"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Channel ID of the message")),
			mcp.WithString("messageId", mcp.Required(), mcp.Description("ID of the message to edit")),
			mcp.WithString("content", mcp.Required(), mcp.Description("New message content")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			messageID := getString(req.Params.Arguments, "messageId")
			content := getString(req.Params.Arguments, "content")
			if channelID == "" || messageID == "" || content == "" {
				return errorResult(fmt.Errorf("channelId, messageId, and content are required")), nil
			}

			msg, err := client.Session.ChannelMessageEdit(channelID, messageID, content)
			if err != nil {
				return errorResult(fmt.Errorf("failed to edit message: %w", err)), nil
			}
			return successResult(fmt.Sprintf("Message updated successfully. ID: `%s`", msg.ID)), nil
		},
	)

	// delete_message
	s.AddTool(
		mcp.NewTool("delete_message",
			mcp.WithDescription("Delete a message from a Discord channel"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Channel ID containing the message")),
			mcp.WithString("messageId", mcp.Required(), mcp.Description("ID of message to delete")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			messageID := getString(req.Params.Arguments, "messageId")
			if channelID == "" || messageID == "" {
				return errorResult(fmt.Errorf("channelId and messageId are required")), nil
			}

			err := client.Session.ChannelMessageDelete(channelID, messageID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to delete message: %w", err)), nil
			}
			return successResult(fmt.Sprintf("Message `%s` deleted successfully.", messageID)), nil
		},
	)

	// read_messages
	s.AddTool(
		mcp.NewTool("read_messages",
			mcp.WithDescription("Read message history from a channel with pagination support"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Channel ID to read messages from")),
			mcp.WithNumber("count", mcp.Description("Number of messages to retrieve (1-100, default 50)")),
			mcp.WithString("before", mcp.Description("Message ID cursor: fetch messages before this ID")),
			mcp.WithString("after", mcp.Description("Message ID cursor: fetch messages after this ID")),
			mcp.WithString("around", mcp.Description("Message ID cursor: fetch messages around this ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			if channelID == "" {
				return errorResult(fmt.Errorf("channelId is required")), nil
			}
			count := getInt(req.Params.Arguments, "count", 50)
			if count < 1 {
				count = 1
			} else if count > 100 {
				count = 100
			}

			before := getString(req.Params.Arguments, "before")
			after := getString(req.Params.Arguments, "after")
			around := getString(req.Params.Arguments, "around")

			messages, err := client.Session.ChannelMessages(channelID, count, before, after, around)
			if err != nil {
				return errorResult(fmt.Errorf("failed to read messages: %w", err)), nil
			}

			if len(messages) == 0 {
				return successResult("No messages found in channel."), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Messages in Channel `%s` (%d returned):\n\n", channelID, len(messages)))
			for i := len(messages) - 1; i >= 0; i-- {
				m := messages[i]
				author := "Unknown"
				if m.Author != nil {
					author = m.Author.Username
				}
				sb.WriteString(fmt.Sprintf("**[%s]** `%s` (%s):\n%s\n",
					m.Timestamp.Format("2006-01-02 15:04:05"),
					m.ID,
					author,
					m.Content,
				))
				if len(m.Attachments) > 0 {
					sb.WriteString(fmt.Sprintf("  *Attachments (%d)*\n", len(m.Attachments)))
				}
				sb.WriteString("\n")
			}
			return successResult(sb.String()), nil
		},
	)

	// add_reaction
	s.AddTool(
		mcp.NewTool("add_reaction",
			mcp.WithDescription("Add an emoji reaction to a message"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Channel ID")),
			mcp.WithString("messageId", mcp.Required(), mcp.Description("Message ID")),
			mcp.WithString("emoji", mcp.Required(), mcp.Description("Unicode emoji (e.g. 👍) or custom format (name:id)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			messageID := getString(req.Params.Arguments, "messageId")
			emoji := getString(req.Params.Arguments, "emoji")
			if channelID == "" || messageID == "" || emoji == "" {
				return errorResult(fmt.Errorf("channelId, messageId, and emoji are required")), nil
			}

			err := client.Session.MessageReactionAdd(channelID, messageID, emoji)
			if err != nil {
				return errorResult(fmt.Errorf("failed to add reaction: %w", err)), nil
			}
			return successResult(fmt.Sprintf("Reaction %s added to message `%s`.", emoji, messageID)), nil
		},
	)

	// remove_reaction
	s.AddTool(
		mcp.NewTool("remove_reaction",
			mcp.WithDescription("Remove bot's reaction from a message"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Channel ID")),
			mcp.WithString("messageId", mcp.Required(), mcp.Description("Message ID")),
			mcp.WithString("emoji", mcp.Required(), mcp.Description("Unicode emoji or custom format to remove")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			messageID := getString(req.Params.Arguments, "messageId")
			emoji := getString(req.Params.Arguments, "emoji")
			if channelID == "" || messageID == "" || emoji == "" {
				return errorResult(fmt.Errorf("channelId, messageId, and emoji are required")), nil
			}

			err := client.Session.MessageReactionRemove(channelID, messageID, emoji, "@me")
			if err != nil {
				return errorResult(fmt.Errorf("failed to remove reaction: %w", err)), nil
			}
			return successResult(fmt.Sprintf("Reaction %s removed from message `%s`.", emoji, messageID)), nil
		},
	)

	// purge_messages
	s.AddTool(
		mcp.NewTool("purge_messages",
			mcp.WithDescription("Bulk delete up to 100 recent messages from a text channel"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Channel ID to purge messages from")),
			mcp.WithNumber("count", mcp.Description("Number of messages to delete (1-100, default 50)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			if channelID == "" {
				return errorResult(fmt.Errorf("channelId is required")), nil
			}

			if err := client.CheckChannelAllowed(channelID); err != nil {
				return errorResult(err), nil
			}

			count := getInt(req.Params.Arguments, "count", 50)
			if count < 1 {
				count = 1
			} else if count > 100 {
				count = 100
			}

			messages, err := client.Session.ChannelMessages(channelID, count, "", "", "")
			if err != nil {
				return errorResult(fmt.Errorf("failed to retrieve messages for purge: %w", err)), nil
			}

			if len(messages) == 0 {
				return successResult("No messages found to purge."), nil
			}

			messageIDs := make([]string, len(messages))
			for i, m := range messages {
				messageIDs[i] = m.ID
			}

			err = client.Session.ChannelMessagesBulkDelete(channelID, messageIDs)
			if err != nil {
				return errorResult(fmt.Errorf("bulk delete failed: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Successfully purged %d messages from channel `%s`.", len(messageIDs), channelID)), nil
		},
	)
}
