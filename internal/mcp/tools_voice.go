package mcp

import (
	"context"
	"fmt"

	"github.com/0xOpCode/discord-mcp-go/internal/discord"
	"github.com/bwmarrin/discordgo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterVoiceTools(s *server.MCPServer, client *discord.Client) {
	// create_voice_channel
	s.AddTool(
		mcp.NewTool("create_voice_channel",
			mcp.WithDescription("Create a new voice channel in a Discord server"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Channel name")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
			mcp.WithString("categoryId", mcp.Description("Parent category ID")),
			mcp.WithNumber("bitrate", mcp.Description("Bitrate in bits (e.g. 64000)")),
			mcp.WithNumber("userLimit", mcp.Description("User limit (0 for unlimited)")),
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

			categoryID := getString(req.Params.Arguments, "categoryId")
			bitrate := getInt(req.Params.Arguments, "bitrate", 64000)
			userLimit := getInt(req.Params.Arguments, "userLimit", 0)

			data := discordgo.GuildChannelCreateData{
				Name:      name,
				Type:      discordgo.ChannelTypeGuildVoice,
				ParentID:  categoryID,
				Bitrate:   bitrate,
				UserLimit: userLimit,
			}

			ch, err := client.Session.GuildChannelCreateComplex(guildID, data)
			if err != nil {
				return errorResult(fmt.Errorf("failed to create voice channel: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Voice channel `%s` created successfully. ID: `%s`", ch.Name, ch.ID)), nil
		},
	)

	// create_stage_channel
	s.AddTool(
		mcp.NewTool("create_stage_channel",
			mcp.WithDescription("Create a new stage channel for audio events in a Discord server"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Channel name")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
			mcp.WithString("categoryId", mcp.Description("Parent category ID")),
			mcp.WithNumber("bitrate", mcp.Description("Bitrate in bits")),
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

			categoryID := getString(req.Params.Arguments, "categoryId")
			bitrate := getInt(req.Params.Arguments, "bitrate", 64000)

			data := discordgo.GuildChannelCreateData{
				Name:     name,
				Type:     discordgo.ChannelTypeGuildStageVoice,
				ParentID: categoryID,
				Bitrate:  bitrate,
			}

			ch, err := client.Session.GuildChannelCreateComplex(guildID, data)
			if err != nil {
				return errorResult(fmt.Errorf("failed to create stage channel: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Stage channel `%s` created successfully. ID: `%s`", ch.Name, ch.ID)), nil
		},
	)

	// edit_voice_channel
	s.AddTool(
		mcp.NewTool("edit_voice_channel",
			mcp.WithDescription("Edit settings of a voice or stage channel"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Voice channel ID")),
			mcp.WithString("name", mcp.Description("New channel name")),
			mcp.WithNumber("bitrate", mcp.Description("Bitrate in bits")),
			mcp.WithNumber("userLimit", mcp.Description("User limit")),
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
			if _, ok := req.Params.Arguments["bitrate"]; ok {
				edit.Bitrate = getInt(req.Params.Arguments, "bitrate", 64000)
			}
			if _, ok := req.Params.Arguments["userLimit"]; ok {
				edit.UserLimit = getInt(req.Params.Arguments, "userLimit", 0)
			}

			ch, err := client.Session.ChannelEditComplex(channelID, edit)
			if err != nil {
				return errorResult(fmt.Errorf("failed to edit voice channel: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Voice channel `%s` updated successfully.", ch.Name)), nil
		},
	)

	// move_member
	s.AddTool(
		mcp.NewTool("move_member",
			mcp.WithDescription("Move a member to another voice channel"),
			mcp.WithString("userId", mcp.Required(), mcp.Description("Target user ID")),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Destination voice channel ID")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			userID := getString(req.Params.Arguments, "userId")
			channelID := getString(req.Params.Arguments, "channelId")
			if userID == "" || channelID == "" {
				return errorResult(fmt.Errorf("userId and channelId are required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			err = client.Session.GuildMemberMove(guildID, userID, &channelID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to move member: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Member `%s` moved to channel `%s`.", userID, channelID)), nil
		},
	)

	// disconnect_member
	s.AddTool(
		mcp.NewTool("disconnect_member",
			mcp.WithDescription("Disconnect a member from their current voice channel"),
			mcp.WithString("userId", mcp.Required(), mcp.Description("Target user ID")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			userID := getString(req.Params.Arguments, "userId")
			if userID == "" {
				return errorResult(fmt.Errorf("userId is required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			err = client.Session.GuildMemberMove(guildID, userID, nil)
			if err != nil {
				return errorResult(fmt.Errorf("failed to disconnect member: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Member `%s` disconnected from voice channel.", userID)), nil
		},
	)

	// modify_voice_state
	s.AddTool(
		mcp.NewTool("modify_voice_state",
			mcp.WithDescription("Server mute or deafen a member in voice channels"),
			mcp.WithString("userId", mcp.Required(), mcp.Description("Target user ID")),
			mcp.WithBoolean("mute", mcp.Description("Server mute")),
			mcp.WithBoolean("deafen", mcp.Description("Server deafen")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			userID := getString(req.Params.Arguments, "userId")
			if userID == "" {
				return errorResult(fmt.Errorf("userId is required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			if _, ok := req.Params.Arguments["mute"]; ok {
				mute := getBool(req.Params.Arguments, "mute", false)
				if err := client.Session.GuildMemberMute(guildID, userID, mute); err != nil {
					return errorResult(fmt.Errorf("failed to set mute: %w", err)), nil
				}
			}

			if _, ok := req.Params.Arguments["deafen"]; ok {
				deafen := getBool(req.Params.Arguments, "deafen", false)
				if err := client.Session.GuildMemberDeafen(guildID, userID, deafen); err != nil {
					return errorResult(fmt.Errorf("failed to set deafen: %w", err)), nil
				}
			}

			return successResult(fmt.Sprintf("Voice state updated for user `%s`.", userID)), nil
		},
	)
}
