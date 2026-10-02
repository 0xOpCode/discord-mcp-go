package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/0xOpCode/discord-mcp-go/internal/discord"
	"github.com/bwmarrin/discordgo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterMessageTools(s *server.MCPServer, client *discord.Client) {
	// send_message
	s.AddTool(
		mcp.NewTool("send_message",
			mcp.WithDescription("Send a message, reply, or rich embeds to a specific Discord text channel"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Channel ID where message will be posted")),
			mcp.WithString("content", mcp.Description("Text message body (optional if embedsJson is provided)")),
			mcp.WithString("replyToMessageId", mcp.Description("Optional parent message ID to reply to")),
			mcp.WithString("embedsJson", mcp.Description("Optional JSON array or single object of Discord embeds")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			content := getString(req.Params.Arguments, "content")
			replyTo := getString(req.Params.Arguments, "replyToMessageId")
			embedsRaw := getString(req.Params.Arguments, "embedsJson")

			if channelID == "" {
				return errorResult(fmt.Errorf("channelId is required")), nil
			}

			if content == "" && embedsRaw == "" {
				return errorResult(fmt.Errorf("content or embedsJson is required")), nil
			}

			if err := client.CheckChannelAllowed(channelID); err != nil {
				return errorResult(err), nil
			}

			msgSend := &discordgo.MessageSend{
				Content: content,
			}

			if replyTo != "" {
				msgSend.Reference = &discordgo.MessageReference{
					MessageID: replyTo,
					ChannelID: channelID,
				}
			}

			if embedsRaw != "" {
				var embeds []*discordgo.MessageEmbed
				if err := json.Unmarshal([]byte(embedsRaw), &embeds); err != nil {
					var single discordgo.MessageEmbed
					if singleErr := json.Unmarshal([]byte(embedsRaw), &single); singleErr == nil {
						embeds = []*discordgo.MessageEmbed{&single}
					} else {
						return errorResult(fmt.Errorf("invalid embedsJson: %w", err)), nil
					}
				}
				msgSend.Embeds = embeds
			}

			msg, err := client.Session.ChannelMessageSendComplex(channelID, msgSend)
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

	// get_message
	s.AddTool(
		mcp.NewTool("get_message",
			mcp.WithDescription("Retrieve a specific Discord message by channel ID and message ID"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Channel ID containing the message")),
			mcp.WithString("messageId", mcp.Required(), mcp.Description("Message ID to fetch")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			messageID := getString(req.Params.Arguments, "messageId")
			if channelID == "" || messageID == "" {
				return errorResult(fmt.Errorf("channelId and messageId are required")), nil
			}

			if err := client.CheckChannelAllowed(channelID); err != nil {
				return errorResult(err), nil
			}

			m, err := client.Session.ChannelMessage(channelID, messageID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to fetch message: %w", err)), nil
			}

			var sb strings.Builder
			authorName := "Unknown"
			authorID := "N/A"
			if m.Author != nil {
				authorName = m.Author.Username
				authorID = m.Author.ID
			}

			sb.WriteString(fmt.Sprintf("### Message Details (`%s`)\n\n", m.ID))
			sb.WriteString(fmt.Sprintf("- **Author:** %s (`%s`)\n", authorName, authorID))
			sb.WriteString(fmt.Sprintf("- **Channel ID:** `%s`\n", m.ChannelID))
			sb.WriteString(fmt.Sprintf("- **Timestamp:** %s\n", m.Timestamp.Format("2006-01-02 15:04:05 UTC")))
			if m.EditedTimestamp != nil {
				sb.WriteString(fmt.Sprintf("- **Edited:** %s\n", m.EditedTimestamp.Format("2006-01-02 15:04:05 UTC")))
			}
			sb.WriteString(fmt.Sprintf("- **Content:**\n%s\n", m.Content))

			if len(m.Attachments) > 0 {
				sb.WriteString(fmt.Sprintf("\n**Attachments (%d):**\n", len(m.Attachments)))
				for _, att := range m.Attachments {
					sb.WriteString(fmt.Sprintf("- [%s](%s) (%d bytes, %s)\n", att.Filename, att.URL, att.Size, att.ContentType))
				}
			}

			if len(m.Reactions) > 0 {
				sb.WriteString("\n**Reactions:**\n")
				for _, r := range m.Reactions {
					sb.WriteString(fmt.Sprintf("- %s: %d\n", r.Emoji.Name, r.Count))
				}
			}

			return successResult(sb.String()), nil
		},
	)

	// send_file
	s.AddTool(
		mcp.NewTool("send_file",
			mcp.WithDescription("Upload a file or attachment to a Discord text channel"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Channel ID where file will be sent")),
			mcp.WithString("filePath", mcp.Required(), mcp.Description("Local file path or base64 data URI")),
			mcp.WithString("content", mcp.Description("Optional message text accompanying file")),
			mcp.WithString("fileName", mcp.Description("Optional custom file name")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			filePath := getString(req.Params.Arguments, "filePath")
			if channelID == "" || filePath == "" {
				return errorResult(fmt.Errorf("channelId and filePath are required")), nil
			}

			if err := client.CheckChannelAllowed(channelID); err != nil {
				return errorResult(err), nil
			}

			customName := getString(req.Params.Arguments, "fileName")
			fileBytes, fileName, err := readFileInput(filePath, customName, client.AllowedFilePaths)
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

			msg, err := client.Session.ChannelMessageSendComplex(channelID, msgSend)
			if err != nil {
				return errorResult(fmt.Errorf("failed to upload file: %w", err)), nil
			}

			return successResult(fmt.Sprintf("File `%s` uploaded to channel `%s`. Message ID: `%s`", fileName, channelID, msg.ID)), nil
		},
	)

	// search_guild_messages
	s.AddTool(
		mcp.NewTool("search_guild_messages",
			mcp.WithDescription("Search messages across channels in a Discord server"),
			mcp.WithString("content", mcp.Description("Text query to search for")),
			mcp.WithString("channelId", mcp.Description("Optional channel ID filter")),
			mcp.WithString("authorId", mcp.Description("Optional author user ID filter")),
			mcp.WithString("mentions", mcp.Description("Optional mentioned user ID filter")),
			mcp.WithString("has", mcp.Description("Filter by attachment type: 'link', 'embed', 'file', 'video', 'image', 'sound', 'sticker'")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
			mcp.WithNumber("limit", mcp.Description("Number of messages (1-25, default 25)")),
			mcp.WithNumber("offset", mcp.Description("Pagination offset (default 0)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			q := url.Values{}
			if content := getString(req.Params.Arguments, "content"); content != "" {
				q.Set("content", content)
			}
			if channelID := getString(req.Params.Arguments, "channelId"); channelID != "" {
				q.Set("channel_id", channelID)
			}
			if authorID := getString(req.Params.Arguments, "authorId"); authorID != "" {
				q.Set("author_id", authorID)
			}
			if mentions := getString(req.Params.Arguments, "mentions"); mentions != "" {
				q.Set("mentions", mentions)
			}
			if has := getString(req.Params.Arguments, "has"); has != "" {
				q.Set("has", has)
			}

			limit := getInt(req.Params.Arguments, "limit", 25)
			if limit < 1 {
				limit = 1
			} else if limit > 25 {
				limit = 25
			}
			q.Set("limit", strconv.Itoa(limit))

			offset := getInt(req.Params.Arguments, "offset", 0)
			if offset > 0 {
				q.Set("offset", strconv.Itoa(offset))
			}

			endpoint := fmt.Sprintf("https://discord.com/api/v10/guilds/%s/messages/search?%s", guildID, q.Encode())
			raw, err := client.Session.Request("GET", endpoint, nil)
			if err != nil {
				return errorResult(fmt.Errorf("failed to search messages: %w", err)), nil
			}

			var searchResp struct {
				TotalResults int                      `json:"total_results"`
				Messages     [][]*discordgo.Message   `json:"messages"`
			}

			if err := json.Unmarshal(raw, &searchResp); err != nil {
				return errorResult(fmt.Errorf("failed to parse search response: %w", err)), nil
			}

			if searchResp.TotalResults == 0 || len(searchResp.Messages) == 0 {
				return successResult(fmt.Sprintf("No messages matched query in server `%s`.", guildID)), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Search Results in Server `%s` (%d total matches):\n\n", guildID, searchResp.TotalResults))
			sb.WriteString(fmt.Sprintf("| %-20s | %-20s | %-20s | %-30s |\n", "Author", "Channel ID", "Message ID", "Content Preview"))
			sb.WriteString("|----------------------|----------------------|----------------------|--------------------------------|\n")

			for _, group := range searchResp.Messages {
				if len(group) == 0 {
					continue
				}
				m := group[0]
				author := "Unknown"
				if m.Author != nil {
					author = m.Author.Username
				}
				if len(author) > 20 {
					author = author[:17] + "..."
				}

				preview := strings.ReplaceAll(m.Content, "\n", " ")
				if len(preview) > 30 {
					preview = preview[:27] + "..."
				}
				if preview == "" && len(m.Attachments) > 0 {
					preview = fmt.Sprintf("[%d attachments]", len(m.Attachments))
				}

				sb.WriteString(fmt.Sprintf("| %-20s | %-20s | %-20s | %-30s |\n", author, m.ChannelID, m.ID, preview))
			}

			return successResult(sb.String()), nil
		},
	)

	// list_pinned_messages
	s.AddTool(
		mcp.NewTool("list_pinned_messages",
			mcp.WithDescription("List all pinned messages in a channel"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Channel ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			if channelID == "" {
				return errorResult(fmt.Errorf("channelId is required")), nil
			}

			if err := client.CheckChannelAllowed(channelID); err != nil {
				return errorResult(err), nil
			}

			messages, err := client.Session.ChannelMessagesPinned(channelID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to fetch pinned messages: %w", err)), nil
			}

			if len(messages) == 0 {
				return successResult("No pinned messages found in channel."), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Pinned Messages in Channel `%s` (%d total):\n\n", channelID, len(messages)))
			sb.WriteString(fmt.Sprintf("| %-20s | %-20s | %-20s | %-30s |\n", "Author", "Message ID", "Timestamp", "Content Preview"))
			sb.WriteString("|----------------------|----------------------|----------------------|--------------------------------|\n")

			for _, m := range messages {
				author := "Unknown"
				if m.Author != nil {
					author = m.Author.Username
				}
				if len(author) > 20 {
					author = author[:17] + "..."
				}

				preview := strings.ReplaceAll(m.Content, "\n", " ")
				if len(preview) > 30 {
					preview = preview[:27] + "..."
				}
				if preview == "" && len(m.Attachments) > 0 {
					preview = fmt.Sprintf("[%d attachments]", len(m.Attachments))
				}

				timestamp := m.Timestamp.Format("2006-01-02 15:04:05")
				sb.WriteString(fmt.Sprintf("| %-20s | %-20s | %-20s | %-30s |\n", author, m.ID, timestamp, preview))
			}

			return successResult(sb.String()), nil
		},
	)

	// pin_message
	s.AddTool(
		mcp.NewTool("pin_message",
			mcp.WithDescription("Pin a message in a channel"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Channel ID")),
			mcp.WithString("messageId", mcp.Required(), mcp.Description("Message ID to pin")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			messageID := getString(req.Params.Arguments, "messageId")
			if channelID == "" || messageID == "" {
				return errorResult(fmt.Errorf("channelId and messageId are required")), nil
			}

			if err := client.CheckChannelAllowed(channelID); err != nil {
				return errorResult(err), nil
			}

			if err := client.Session.ChannelMessagePin(channelID, messageID); err != nil {
				return errorResult(fmt.Errorf("failed to pin message: %w", err)), nil
			}
			return successResult(fmt.Sprintf("Message `%s` pinned in channel `%s`.", messageID, channelID)), nil
		},
	)

	// unpin_message
	s.AddTool(
		mcp.NewTool("unpin_message",
			mcp.WithDescription("Unpin a message from a channel"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Channel ID")),
			mcp.WithString("messageId", mcp.Required(), mcp.Description("Message ID to unpin")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			messageID := getString(req.Params.Arguments, "messageId")
			if channelID == "" || messageID == "" {
				return errorResult(fmt.Errorf("channelId and messageId are required")), nil
			}

			if err := client.CheckChannelAllowed(channelID); err != nil {
				return errorResult(err), nil
			}

			if err := client.Session.ChannelMessageUnpin(channelID, messageID); err != nil {
				return errorResult(fmt.Errorf("failed to unpin message: %w", err)), nil
			}
			return successResult(fmt.Sprintf("Message `%s` unpinned from channel `%s`.", messageID, channelID)), nil
		},
	)

	// crosspost_message
	s.AddTool(
		mcp.NewTool("crosspost_message",
			mcp.WithDescription("Publish an announcement message to all follower channels"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Channel ID containing announcement")),
			mcp.WithString("messageId", mcp.Required(), mcp.Description("Message ID to crosspost")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			messageID := getString(req.Params.Arguments, "messageId")
			if channelID == "" || messageID == "" {
				return errorResult(fmt.Errorf("channelId and messageId are required")), nil
			}

			if err := client.CheckChannelAllowed(channelID); err != nil {
				return errorResult(err), nil
			}

			msg, err := client.Session.ChannelMessageCrosspost(channelID, messageID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to crosspost message: %w", err)), nil
			}
			return successResult(fmt.Sprintf("Message `%s` published to follower channels.", msg.ID)), nil
		},
	)
}
