package mcp

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/mark3labs/mcp-go/mcp"
)

func readFileInput(input, defaultName string, allowedPaths []string) ([]byte, string, error) {
	cleanInput := strings.TrimSpace(input)
	if cleanInput == "" {
		return nil, "", fmt.Errorf("file path or base64 data required")
	}

	if strings.HasPrefix(cleanInput, "data:") {
		idx := strings.Index(cleanInput, ",")
		if idx == -1 {
			return nil, "", fmt.Errorf("malformed data URI")
		}
		header := cleanInput[:idx]
		data := cleanInput[idx+1:]
		ext := "bin"
		if strings.Contains(header, "image/png") {
			ext = "png"
		} else if strings.Contains(header, "image/jpeg") || strings.Contains(header, "image/jpg") {
			ext = "jpg"
		} else if strings.Contains(header, "image/gif") {
			ext = "gif"
		} else if strings.Contains(header, "application/pdf") {
			ext = "pdf"
		} else if strings.Contains(header, "text/plain") {
			ext = "txt"
		} else if strings.Contains(header, "application/json") {
			ext = "json"
		}

		raw, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			return nil, "", fmt.Errorf("invalid base64 in data URI: %w", err)
		}

		name := defaultName
		if name == "" {
			name = "attachment." + ext
		}
		return raw, name, nil
	}

	if len(allowedPaths) > 0 {
		absPath, err := filepath.Abs(cleanInput)
		if err != nil {
			return nil, "", fmt.Errorf("invalid path: %w", err)
		}
		absPath = filepath.Clean(absPath)

		allowed := false
		for _, dir := range allowedPaths {
			cleanDir := filepath.Clean(dir)
			rel, relErr := filepath.Rel(cleanDir, absPath)
			if relErr == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				allowed = true
				break
			}
		}

		if !allowed {
			if !strings.Contains(cleanInput, "/") && !strings.Contains(cleanInput, "\\") {
				if raw, err := base64.StdEncoding.DecodeString(cleanInput); err == nil {
					name := defaultName
					if name == "" {
						name = "attachment.bin"
					}
					return raw, name, nil
				}
			}
			return nil, "", fmt.Errorf("access to file '%s' prohibited by allowed path configuration", cleanInput)
		}
	}

	if data, err := os.ReadFile(cleanInput); err == nil {
		name := defaultName
		if name == "" {
			name = filepath.Base(cleanInput)
		}
		return data, name, nil
	}

	if raw, err := base64.StdEncoding.DecodeString(cleanInput); err == nil {
		name := defaultName
		if name == "" {
			name = "attachment.bin"
		}
		return raw, name, nil
	}

	return nil, "", fmt.Errorf("file not found on disk and input is not valid base64")
}

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
