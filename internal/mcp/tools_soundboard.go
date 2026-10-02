package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/0xOpCode/discord-mcp-go/internal/discord"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type soundboardSound struct {
	SoundID   string  `json:"sound_id"`
	Name      string  `json:"name"`
	Volume    float64 `json:"volume"`
	EmojiName string  `json:"emoji_name,omitempty"`
}

type soundboardResponse struct {
	Items []soundboardSound `json:"items"`
}

func RegisterSoundboardTools(s *server.MCPServer, client *discord.Client) {
	// list_soundboard_sounds
	s.AddTool(
		mcp.NewTool("list_soundboard_sounds",
			mcp.WithDescription("List custom soundboard audio clips in a Discord server"),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			endpoint := fmt.Sprintf("https://discord.com/api/v10/guilds/%s/soundboard-sounds", guildID)
			raw, err := client.Session.Request("GET", endpoint, nil)
			if err != nil {
				return errorResult(fmt.Errorf("failed to fetch soundboard sounds: %w", err)), nil
			}

			var sounds []soundboardSound
			if err := json.Unmarshal(raw, &sounds); err != nil {
				var wrapped soundboardResponse
				if errWrap := json.Unmarshal(raw, &wrapped); errWrap == nil {
					sounds = wrapped.Items
				}
			}

			if len(sounds) == 0 {
				return successResult(fmt.Sprintf("No custom soundboard sounds found in server `%s`.", guildID)), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Soundboard Sounds in Server `%s` (%d total):\n\n", guildID, len(sounds)))
			sb.WriteString(fmt.Sprintf("| %-25s | %-20s | %-8s |\n", "Sound Name", "Sound ID", "Volume"))
			sb.WriteString("|---------------------------|----------------------|----------|\n")
			for _, snd := range sounds {
				name := snd.Name
				if len(name) > 25 {
					name = name[:22] + "..."
				}
				sb.WriteString(fmt.Sprintf("| %-25s | %-20s | %-8.2f |\n", name, snd.SoundID, snd.Volume))
			}
			return successResult(sb.String()), nil
		},
	)

	// create_soundboard_sound
	s.AddTool(
		mcp.NewTool("create_soundboard_sound",
			mcp.WithDescription("Upload a new custom soundboard sound to a Discord server"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Sound name")),
			mcp.WithString("sound", mcp.Required(), mcp.Description("Base64 audio data URI (e.g. 'data:audio/mp3;base64,...')")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
			mcp.WithNumber("volume", mcp.Description("Volume level from 0.0 to 1.0 (default: 1.0)")),
			mcp.WithString("emojiName", mcp.Description("Optional emoji name linked to the sound")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name := getString(req.Params.Arguments, "name")
			soundData := getString(req.Params.Arguments, "sound")
			if name == "" || soundData == "" {
				return errorResult(fmt.Errorf("name and sound are required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			payload := map[string]interface{}{
				"name":  name,
				"sound": soundData,
			}
			if _, ok := req.Params.Arguments["volume"]; ok {
				if v, ok := req.Params.Arguments["volume"].(float64); ok {
					payload["volume"] = v
				}
			}
			if emoji := getString(req.Params.Arguments, "emojiName"); emoji != "" {
				payload["emoji_name"] = emoji
			}

			body, _ := json.Marshal(payload)
			endpoint := fmt.Sprintf("https://discord.com/api/v10/guilds/%s/soundboard-sounds", guildID)
			respBytes, err := client.Session.Request("POST", endpoint, bytes.NewReader(body))
			if err != nil {
				return errorResult(fmt.Errorf("failed to create soundboard sound: %w", err)), nil
			}

			var created soundboardSound
			_ = json.Unmarshal(respBytes, &created)
			return successResult(fmt.Sprintf("Soundboard sound `%s` created. Sound ID: `%s`", created.Name, created.SoundID)), nil
		},
	)

	// delete_soundboard_sound
	s.AddTool(
		mcp.NewTool("delete_soundboard_sound",
			mcp.WithDescription("Delete a custom soundboard sound from a server"),
			mcp.WithString("soundId", mcp.Required(), mcp.Description("Sound ID to delete")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			soundID := getString(req.Params.Arguments, "soundId")
			if soundID == "" {
				return errorResult(fmt.Errorf("soundId is required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			endpoint := fmt.Sprintf("https://discord.com/api/v10/guilds/%s/soundboard-sounds/%s", guildID, soundID)
			_, err = client.Session.Request("DELETE", endpoint, nil)
			if err != nil {
				return errorResult(fmt.Errorf("failed to delete soundboard sound: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Soundboard sound `%s` deleted from server `%s`.", soundID, guildID)), nil
		},
	)
}
