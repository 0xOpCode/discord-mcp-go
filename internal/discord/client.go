package discord

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Client struct {
	Session          *discordgo.Session
	DefaultGuildID   string
	BotUser          *discordgo.User
	Blacklist        map[string]struct{}
	AllowedFilePaths []string
	OwnerUserID      string
}

func NewClient(token, defaultGuildID string, blacklistedGuilds []string) (*Client, error) {
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

	// Open gateway session for real-time presence/activity updates
	_ = session.Open()

	blacklistMap := make(map[string]struct{})
	for _, id := range blacklistedGuilds {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			blacklistMap[trimmed] = struct{}{}
		}
	}

	return &Client{
		Session:        session,
		DefaultGuildID: defaultGuildID,
		BotUser:        botUser,
		Blacklist:      blacklistMap,
	}, nil
}

func (c *Client) IsBlacklisted(guildID string) bool {
	if guildID == "" {
		return false
	}
	_, found := c.Blacklist[guildID]
	return found
}

func (c *Client) IsPathAllowed(filePath string) error {
	if c == nil || len(c.AllowedFilePaths) == 0 {
		return nil
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	absPath = filepath.Clean(absPath)

	for _, allowed := range c.AllowedFilePaths {
		cleanAllowed := filepath.Clean(allowed)
		rel, err := filepath.Rel(cleanAllowed, absPath)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return nil
		}
	}

	return fmt.Errorf("access to file '%s' prohibited by allowed path configuration", filePath)
}

func (c *Client) CheckGuildAllowed(guildID string) error {
	if c.IsBlacklisted(guildID) {
		return fmt.Errorf("action blocked: server %s is blacklisted", guildID)
	}
	return nil
}

func (c *Client) CheckChannelAllowed(channelID string) error {
	if channelID == "" {
		return nil
	}
	ch, err := c.Session.Channel(channelID)
	if err != nil {
		return nil
	}
	if ch.GuildID != "" && c.IsBlacklisted(ch.GuildID) {
		return fmt.Errorf("action blocked: channel %s belongs to blacklisted server %s", channelID, ch.GuildID)
	}
	return nil
}

func (c *Client) ResolveGuildID(override string) (string, error) {
	target := strings.TrimSpace(override)
	if target == "" {
		target = strings.TrimSpace(c.DefaultGuildID)
	}
	if target == "" {
		return "", fmt.Errorf("guild_id parameter is required when DISCORD_GUILD_ID is not configured")
	}
	if err := c.CheckGuildAllowed(target); err != nil {
		return "", err
	}
	return target, nil
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
