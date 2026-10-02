package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/0xOpCode/discord-mcp-go/internal/discord"
	"github.com/0xOpCode/discord-mcp-go/internal/storage"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type GuildOverviewResource struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	OwnerID     string            `json:"ownerId"`
	MemberCount int               `json:"memberCount"`
	Description string            `json:"description,omitempty"`
	Channels    []ChannelSummary  `json:"channels"`
}

type ChannelSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type int    `json:"type"`
}

type MessageSummary struct {
	ID          string   `json:"id"`
	Author      string   `json:"author"`
	AuthorID    string   `json:"authorId"`
	Content     string   `json:"content"`
	Timestamp   string   `json:"timestamp"`
	Attachments []string `json:"attachments,omitempty"`
}

type RoleSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Color       int    `json:"color"`
	Position    int    `json:"position"`
	Permissions int64  `json:"permissions"`
	Mentionable bool   `json:"mentionable"`
}

func HandleMemoryRead(store *storage.Store) server.ResourceHandlerFunc {
	return func(ctx context.Context, req mcp.ReadResourceRequest) ([]interface{}, error) {
		if store == nil {
			return nil, fmt.Errorf("persistent store not available")
		}
		key := strings.TrimPrefix(req.Params.URI, "discord://memory/")
		if key == "" || key == req.Params.URI {
			return nil, fmt.Errorf("invalid memory uri: %s", req.Params.URI)
		}

		item, err := store.Get(key)
		if err != nil {
			return nil, err
		}

		mimeType := item.MimeType
		if mimeType == "" {
			mimeType = "text/plain"
		}

		return []interface{}{
			mcp.TextResourceContents{
				ResourceContents: mcp.ResourceContents{
					URI:      req.Params.URI,
					MIMEType: mimeType,
				},
				Text: item.Content,
			},
		}, nil
	}
}

func RegisterResources(s *server.MCPServer, client *discord.Client, store *storage.Store) {
	// 1. discord://memory/{key}
	memoryHandler := HandleMemoryRead(store)

	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"discord://memory/{key}",
			"Persistent Memory Document",
			mcp.WithTemplateDescription("Persistent document stored under a key in the local store"),
			mcp.WithTemplateMIMEType("text/plain"),
		),
		server.ResourceTemplateHandlerFunc(memoryHandler),
	)

	// 2. discord://guilds/{guildId}/overview
	guildOverviewHandler := func(ctx context.Context, req mcp.ReadResourceRequest) ([]interface{}, error) {
		if client == nil || client.Session == nil {
			return nil, fmt.Errorf("discord client not connected")
		}

		uri := req.Params.URI
		if !strings.HasPrefix(uri, "discord://guilds/") || !strings.HasSuffix(uri, "/overview") {
			return nil, fmt.Errorf("invalid guild overview uri: %s", uri)
		}
		guildID := strings.TrimSuffix(strings.TrimPrefix(uri, "discord://guilds/"), "/overview")
		if guildID == "" {
			return nil, fmt.Errorf("missing guild id in uri: %s", uri)
		}

		if err := client.CheckGuildAllowed(guildID); err != nil {
			return nil, err
		}

		guild, err := client.Session.Guild(guildID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch guild: %w", err)
		}

		channels, err := client.Session.GuildChannels(guildID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch channels: %w", err)
		}

		channelSummaries := make([]ChannelSummary, 0, len(channels))
		for _, ch := range channels {
			channelSummaries = append(channelSummaries, ChannelSummary{
				ID:   ch.ID,
				Name: ch.Name,
				Type: int(ch.Type),
			})
		}

		overview := GuildOverviewResource{
			ID:          guild.ID,
			Name:        guild.Name,
			OwnerID:     guild.OwnerID,
			MemberCount: guild.MemberCount,
			Description: guild.Description,
			Channels:    channelSummaries,
		}

		bytes, err := json.MarshalIndent(overview, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to serialize guild overview: %w", err)
		}

		return []interface{}{
			mcp.TextResourceContents{
				ResourceContents: mcp.ResourceContents{
					URI:      uri,
					MIMEType: "application/json",
				},
				Text: string(bytes),
			},
		}, nil
	}

	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"discord://guilds/{guildId}/overview",
			"Guild Overview",
			mcp.WithTemplateDescription("Server details and list of channels"),
			mcp.WithTemplateMIMEType("application/json"),
		),
		guildOverviewHandler,
	)

	// 3. discord://channels/{channelId}/pinned
	channelPinnedHandler := func(ctx context.Context, req mcp.ReadResourceRequest) ([]interface{}, error) {
		if client == nil || client.Session == nil {
			return nil, fmt.Errorf("discord client not connected")
		}

		uri := req.Params.URI
		if !strings.HasPrefix(uri, "discord://channels/") || !strings.HasSuffix(uri, "/pinned") {
			return nil, fmt.Errorf("invalid pinned messages uri: %s", uri)
		}
		channelID := strings.TrimSuffix(strings.TrimPrefix(uri, "discord://channels/"), "/pinned")
		if channelID == "" {
			return nil, fmt.Errorf("missing channel id in uri: %s", uri)
		}

		if err := client.CheckChannelAllowed(channelID); err != nil {
			return nil, err
		}

		messages, err := client.Session.ChannelMessagesPinned(channelID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch pinned messages: %w", err)
		}

		summaries := make([]MessageSummary, 0, len(messages))
		for _, msg := range messages {
			var atts []string
			for _, a := range msg.Attachments {
				atts = append(atts, a.URL)
			}
			summaries = append(summaries, MessageSummary{
				ID:          msg.ID,
				Author:      msg.Author.Username,
				AuthorID:    msg.Author.ID,
				Content:     msg.Content,
				Timestamp:   msg.Timestamp.Format("2006-01-02 15:04:05 UTC"),
				Attachments: atts,
			})
		}

		bytes, err := json.MarshalIndent(summaries, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to serialize pinned messages: %w", err)
		}

		return []interface{}{
			mcp.TextResourceContents{
				ResourceContents: mcp.ResourceContents{
					URI:      uri,
					MIMEType: "application/json",
				},
				Text: string(bytes),
			},
		}, nil
	}

	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"discord://channels/{channelId}/pinned",
			"Channel Pinned Messages",
			mcp.WithTemplateDescription("List of pinned messages in a Discord channel"),
			mcp.WithTemplateMIMEType("application/json"),
		),
		channelPinnedHandler,
	)

	// 4. discord://channels/{channelId}/recent
	channelRecentHandler := func(ctx context.Context, req mcp.ReadResourceRequest) ([]interface{}, error) {
		if client == nil || client.Session == nil {
			return nil, fmt.Errorf("discord client not connected")
		}

		uri := req.Params.URI
		if !strings.HasPrefix(uri, "discord://channels/") || !strings.HasSuffix(uri, "/recent") {
			return nil, fmt.Errorf("invalid recent messages uri: %s", uri)
		}
		channelID := strings.TrimSuffix(strings.TrimPrefix(uri, "discord://channels/"), "/recent")
		if channelID == "" {
			return nil, fmt.Errorf("missing channel id in uri: %s", uri)
		}

		if err := client.CheckChannelAllowed(channelID); err != nil {
			return nil, err
		}

		messages, err := client.Session.ChannelMessages(channelID, 50, "", "", "")
		if err != nil {
			return nil, fmt.Errorf("failed to fetch recent messages: %w", err)
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# Recent Messages in Channel %s (%d messages)\n\n", channelID, len(messages)))

		// Render chronological order (oldest to newest)
		for i := len(messages) - 1; i >= 0; i-- {
			msg := messages[i]
			t := msg.Timestamp.Format("2006-01-02 15:04:05")
			sb.WriteString(fmt.Sprintf("- **[%s] %s**: %s\n", t, msg.Author.Username, msg.Content))
			for _, att := range msg.Attachments {
				sb.WriteString(fmt.Sprintf("  - *Attachment*: %s\n", att.URL))
			}
		}

		return []interface{}{
			mcp.TextResourceContents{
				ResourceContents: mcp.ResourceContents{
					URI:      uri,
					MIMEType: "text/markdown",
				},
				Text: sb.String(),
			},
		}, nil
	}

	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"discord://channels/{channelId}/recent",
			"Channel Recent Messages",
			mcp.WithTemplateDescription("Recent 50 messages from a Discord channel formatted as markdown"),
			mcp.WithTemplateMIMEType("text/markdown"),
		),
		channelRecentHandler,
	)

	// 5. discord://guilds/{guildId}/roles
	guildRolesHandler := func(ctx context.Context, req mcp.ReadResourceRequest) ([]interface{}, error) {
		if client == nil || client.Session == nil {
			return nil, fmt.Errorf("discord client not connected")
		}

		uri := req.Params.URI
		if !strings.HasPrefix(uri, "discord://guilds/") || !strings.HasSuffix(uri, "/roles") {
			return nil, fmt.Errorf("invalid guild roles uri: %s", uri)
		}
		guildID := strings.TrimSuffix(strings.TrimPrefix(uri, "discord://guilds/"), "/roles")
		if guildID == "" {
			return nil, fmt.Errorf("missing guild id in uri: %s", uri)
		}

		if err := client.CheckGuildAllowed(guildID); err != nil {
			return nil, err
		}

		roles, err := client.Session.GuildRoles(guildID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch guild roles: %w", err)
		}

		roleSummaries := make([]RoleSummary, 0, len(roles))
		for _, r := range roles {
			roleSummaries = append(roleSummaries, RoleSummary{
				ID:          r.ID,
				Name:        r.Name,
				Color:       r.Color,
				Position:    r.Position,
				Permissions: r.Permissions,
				Mentionable: r.Mentionable,
			})
		}

		bytes, err := json.MarshalIndent(roleSummaries, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to serialize roles: %w", err)
		}

		return []interface{}{
			mcp.TextResourceContents{
				ResourceContents: mcp.ResourceContents{
					URI:      uri,
					MIMEType: "application/json",
				},
				Text: string(bytes),
			},
		}, nil
	}

	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"discord://guilds/{guildId}/roles",
			"Guild Roles",
			mcp.WithTemplateDescription("List of roles and permissions in a server"),
			mcp.WithTemplateMIMEType("application/json"),
		),
		guildRolesHandler,
	)

	// Register existing persistent resources as concrete entries for resources/list
	if store != nil {
		for _, item := range store.List() {
			RegisterMemoryResource(s, item.Key, item.Description, item.MimeType, memoryHandler)
		}
	}

	// Register default server resources if guild ID is configured
	if client != nil && client.DefaultGuildID != "" {
		s.AddResource(
			mcp.NewResource(
				fmt.Sprintf("discord://guilds/%s/overview", client.DefaultGuildID),
				"Default Guild Overview",
				mcp.WithResourceDescription("Channels and summary for configured default guild"),
				mcp.WithMIMEType("application/json"),
			),
			guildOverviewHandler,
		)
		s.AddResource(
			mcp.NewResource(
				fmt.Sprintf("discord://guilds/%s/roles", client.DefaultGuildID),
				"Default Guild Roles",
				mcp.WithResourceDescription("Roles and permissions for configured default guild"),
				mcp.WithMIMEType("application/json"),
			),
			guildRolesHandler,
		)
	}
}

func RegisterMemoryResource(s *server.MCPServer, key, description, mimeType string, handler server.ResourceHandlerFunc) {
	if mimeType == "" {
		mimeType = "text/plain"
	}
	s.AddResource(
		mcp.NewResource(
			fmt.Sprintf("discord://memory/%s", key),
			fmt.Sprintf("Memory: %s", key),
			mcp.WithResourceDescription(description),
			mcp.WithMIMEType(mimeType),
		),
		handler,
	)
}
