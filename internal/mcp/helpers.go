package mcp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/mark3labs/mcp-go/mcp"
)

func getString(args map[string]interface{}, key string) string {
	if val, ok := args[key]; ok && val != nil {
		if s, ok := val.(string); ok {
			return strings.TrimSpace(s)
		}
		return strings.TrimSpace(fmt.Sprintf("%v", val))
	}
	return ""
}

func getInt(args map[string]interface{}, key string, def int) int {
	if val, ok := args[key]; ok && val != nil {
		switch v := val.(type) {
		case float64:
			return int(v)
		case int:
			return v
		case int64:
			return int(v)
		case string:
			if parsed, err := strconv.Atoi(v); err == nil {
				return parsed
			}
		}
	}
	return def
}

func getBool(args map[string]interface{}, key string, def bool) bool {
	if val, ok := args[key]; ok && val != nil {
		switch v := val.(type) {
		case bool:
			return v
		case string:
			lower := strings.ToLower(strings.TrimSpace(v))
			if lower == "true" || lower == "1" || lower == "yes" {
				return true
			}
			if lower == "false" || lower == "0" || lower == "no" {
				return false
			}
		}
	}
	return def
}

func successResult(text string) *mcp.CallToolResult {
	return mcp.NewToolResultText(text)
}

func errorResult(err error) *mcp.CallToolResult {
	return mcp.NewToolResultError(err.Error())
}

func channelTypeToString(t discordgo.ChannelType) string {
	switch t {
	case discordgo.ChannelTypeGuildText:
		return "Text"
	case discordgo.ChannelTypeDM:
		return "DM"
	case discordgo.ChannelTypeGuildVoice:
		return "Voice"
	case discordgo.ChannelTypeGroupDM:
		return "GroupDM"
	case discordgo.ChannelTypeGuildCategory:
		return "Category"
	case discordgo.ChannelTypeGuildNews:
		return "Announcement"
	case discordgo.ChannelTypeGuildNewsThread:
		return "AnnouncementThread"
	case discordgo.ChannelTypeGuildPublicThread:
		return "PublicThread"
	case discordgo.ChannelTypeGuildPrivateThread:
		return "PrivateThread"
	case discordgo.ChannelTypeGuildStageVoice:
		return "StageVoice"
	case discordgo.ChannelTypeGuildForum:
		return "Forum"
	default:
		return fmt.Sprintf("Type(%d)", t)
	}
}
