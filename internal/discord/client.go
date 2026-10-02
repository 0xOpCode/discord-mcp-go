package discord

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Client struct {
	Session        *discordgo.Session
	DefaultGuildID string
	BotUser        *discordgo.User
}

func NewClient(token, defaultGuildID string) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("DISCORD_TOKEN environment variable is not set")
	}

	if !strings.HasPrefix(token, "Bot ") {
		token = "Bot " + token
	}

	session, err := discordgo.New(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create discord session: %w", err)
	}

	botUser, err := session.User("@me")
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with Discord API: %w", err)
	}

	return &Client{
		Session:        session,
		DefaultGuildID: defaultGuildID,
		BotUser:        botUser,
	}, nil
}

func (c *Client) ResolveGuildID(override string) (string, error) {
	if strings.TrimSpace(override) != "" {
		return strings.TrimSpace(override), nil
	}
	if strings.TrimSpace(c.DefaultGuildID) != "" {
		return strings.TrimSpace(c.DefaultGuildID), nil
	}
	return "", fmt.Errorf("guild_id parameter is required when DISCORD_GUILD_ID is not configured")
}

func (c *Client) ListServers() ([]*discordgo.UserGuild, error) {
	guilds, err := c.Session.UserGuilds(100, "", "", false)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bot guilds: %w", err)
	}
	return guilds, nil
}

func (c *Client) GetServerInfo(guildID string) (*discordgo.Guild, error) {
	targetGuildID, err := c.ResolveGuildID(guildID)
	if err != nil {
		return nil, err
	}
	guild, err := c.Session.Guild(targetGuildID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch guild %s: %w", targetGuildID, err)
	}
	return guild, nil
}

func (c *Client) CheckBotPermissions(guildID string) (map[string]bool, error) {
	targetGuildID, err := c.ResolveGuildID(guildID)
	if err != nil {
		return nil, err
	}

	member, err := c.Session.GuildMember(targetGuildID, c.BotUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bot member in guild %s: %w", targetGuildID, err)
	}

	guild, err := c.Session.Guild(targetGuildID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch guild roles: %w", err)
	}

	roleMap := make(map[string]*discordgo.Role)
	for _, role := range guild.Roles {
		roleMap[role.ID] = role
	}

	var combinedPerms int64
	for _, roleID := range member.Roles {
		if role, exists := roleMap[roleID]; exists {
			combinedPerms |= int64(role.Permissions)
		}
	}
	if everyoneRole, exists := roleMap[targetGuildID]; exists {
		combinedPerms |= int64(everyoneRole.Permissions)
	}

	hasPerm := func(permission int64) bool {
		if (combinedPerms & int64(discordgo.PermissionAdministrator)) != 0 {
			return true
		}
		return (combinedPerms & permission) != 0
	}

	return map[string]bool{
		"Administrator":         hasPerm(int64(discordgo.PermissionAdministrator)),
		"ManageServer":          hasPerm(int64(discordgo.PermissionManageServer)),
		"ManageRoles":           hasPerm(int64(discordgo.PermissionManageRoles)),
		"ManageChannels":        hasPerm(int64(discordgo.PermissionManageChannels)),
		"KickMembers":           hasPerm(int64(discordgo.PermissionKickMembers)),
		"BanMembers":            hasPerm(int64(discordgo.PermissionBanMembers)),
		"SendMessages":          hasPerm(int64(discordgo.PermissionSendMessages)),
		"ManageMessages":        hasPerm(int64(discordgo.PermissionManageMessages)),
		"EmbedLinks":            hasPerm(int64(discordgo.PermissionEmbedLinks)),
		"AttachFiles":           hasPerm(int64(discordgo.PermissionAttachFiles)),
		"ReadMessageHistory":    hasPerm(int64(discordgo.PermissionReadMessageHistory)),
		"ManageWebhooks":        hasPerm(int64(discordgo.PermissionManageWebhooks)),
		"ManageEmojisAndStickers": hasPerm(int64(discordgo.PermissionManageGuildExpressions)),
		"ModerateMembers":       hasPerm(int64(discordgo.PermissionModerateMembers)),
	}, nil
}
