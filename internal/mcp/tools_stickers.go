package mcp

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"os"
	"strings"

	"github.com/0xOpCode/discord-mcp-go/internal/discord"
	"github.com/bwmarrin/discordgo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterStickerTools(s *server.MCPServer, client *discord.Client) {
	// list_stickers
	s.AddTool(
		mcp.NewTool("list_stickers",
			mcp.WithDescription("List custom stickers in a Discord server"),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			endpoint := fmt.Sprintf("https://discord.com/api/v10/guilds/%s/stickers", guildID)
			raw, err := client.Session.Request("GET", endpoint, nil)
			if err != nil {
				return errorResult(fmt.Errorf("failed to fetch stickers: %w", err)), nil
			}

			var stickers []*discordgo.Sticker
			if err := json.Unmarshal(raw, &stickers); err != nil {
				return errorResult(fmt.Errorf("failed to parse stickers response: %w", err)), nil
			}

			if len(stickers) == 0 {
				return successResult(fmt.Sprintf("No custom stickers found in server `%s`.", guildID)), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Custom Stickers in Server `%s` (%d total):\n\n", guildID, len(stickers)))
			sb.WriteString(fmt.Sprintf("| %-22s | %-20s | %-12s | %-25s |\n", "Sticker Name", "Sticker ID", "Format", "Description"))
			sb.WriteString("|------------------------|----------------------|--------------|---------------------------|\n")
			for _, stk := range stickers {
				name := stk.Name
				if len(name) > 22 {
					name = name[:19] + "..."
				}
				desc := stk.Description
				if len(desc) > 25 {
					desc = desc[:22] + "..."
				}
				fmtType := formatTypeName(stk.FormatType)
				sb.WriteString(fmt.Sprintf("| %-22s | %-20s | %-12s | %-25s |\n", name, stk.ID, fmtType, desc))
			}
			return successResult(sb.String()), nil
		},
	)

	// create_sticker
	s.AddTool(
		mcp.NewTool("create_sticker",
			mcp.WithDescription("Upload a new custom sticker to a Discord server"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Sticker name (2-30 characters)")),
			mcp.WithString("tags", mcp.Required(), mcp.Description("Autocomplete tags or related emoji (e.g. 'smile, happy')")),
			mcp.WithString("file", mcp.Required(), mcp.Description("Sticker image: base64 data URI, raw base64 string, or local file path (PNG, APNG, GIF, Lottie)")),
			mcp.WithString("description", mcp.Description("Optional sticker description (2-100 characters)")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name := getString(req.Params.Arguments, "name")
			tags := getString(req.Params.Arguments, "tags")
			fileInput := getString(req.Params.Arguments, "file")
			description := getString(req.Params.Arguments, "description")

			if name == "" || tags == "" || fileInput == "" {
				return errorResult(fmt.Errorf("name, tags, and file are required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			fileBytes, fileName, err := decodeStickerFile(fileInput)
			if err != nil {
				return errorResult(fmt.Errorf("invalid sticker file: %w", err)), nil
			}

			var body bytes.Buffer
			writer := multipart.NewWriter(&body)

			if err := writer.WriteField("name", name); err != nil {
				return errorResult(err), nil
			}
			if err := writer.WriteField("tags", tags); err != nil {
				return errorResult(err), nil
			}
			if description != "" {
				if err := writer.WriteField("description", description); err != nil {
					return errorResult(err), nil
				}
			}

			part, err := writer.CreateFormFile("file", fileName)
			if err != nil {
				return errorResult(err), nil
			}
			if _, err := part.Write(fileBytes); err != nil {
				return errorResult(err), nil
			}

			if err := writer.Close(); err != nil {
				return errorResult(err), nil
			}

			endpoint := fmt.Sprintf("https://discord.com/api/v10/guilds/%s/stickers", guildID)
			respBytes, err := client.Session.Request(
				"POST",
				endpoint,
				&body,
				discordgo.WithHeader("Content-Type", writer.FormDataContentType()),
			)
			if err != nil {
				return errorResult(fmt.Errorf("failed to create sticker: %w", err)), nil
			}

			var created discordgo.Sticker
			if err := json.Unmarshal(respBytes, &created); err != nil {
				return successResult(fmt.Sprintf("Sticker `%s` uploaded to server `%s`.", name, guildID)), nil
			}

			return successResult(fmt.Sprintf("Sticker `%s` created. Sticker ID: `%s` in server `%s`.", created.Name, created.ID, guildID)), nil
		},
	)

	// delete_sticker
	s.AddTool(
		mcp.NewTool("delete_sticker",
			mcp.WithDescription("Delete a custom sticker from a Discord server"),
			mcp.WithString("stickerId", mcp.Required(), mcp.Description("Sticker ID to delete")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			stickerID := getString(req.Params.Arguments, "stickerId")
			if stickerID == "" {
				return errorResult(fmt.Errorf("stickerId is required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			endpoint := fmt.Sprintf("https://discord.com/api/v10/guilds/%s/stickers/%s", guildID, stickerID)
			_, err = client.Session.Request("DELETE", endpoint, nil)
			if err != nil {
				return errorResult(fmt.Errorf("failed to delete sticker: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Sticker `%s` deleted from server `%s`.", stickerID, guildID)), nil
		},
	)
}

func decodeStickerFile(input string) ([]byte, string, error) {
	if strings.HasPrefix(input, "data:") {
		idx := strings.Index(input, ",")
		if idx == -1 {
			return nil, "", fmt.Errorf("malformed data URI")
		}
		header := input[:idx]
		data := input[idx+1:]
		ext := "png"
		if strings.Contains(header, "image/gif") {
			ext = "gif"
		} else if strings.Contains(header, "image/apng") {
			ext = "png"
		} else if strings.Contains(header, "application/json") {
			ext = "json"
		}
		raw, err := base64.StdEncoding.DecodeString(data)
		return raw, "sticker." + ext, err
	}

	if data, err := os.ReadFile(input); err == nil {
		parts := strings.Split(input, "/")
		return data, parts[len(parts)-1], nil
	}

	raw, err := base64.StdEncoding.DecodeString(input)
	if err == nil {
		return raw, "sticker.png", nil
	}

	return nil, "", fmt.Errorf("input must be base64 data URI, raw base64 string, or local file path")
}

func formatTypeName(t discordgo.StickerFormat) string {
	switch t {
	case discordgo.StickerFormatTypePNG:
		return "PNG"
	case discordgo.StickerFormatTypeAPNG:
		return "APNG"
	case discordgo.StickerFormatTypeLottie:
		return "Lottie"
	case discordgo.StickerFormatTypeGIF:
		return "GIF"
	default:
		return "Unknown"
	}
}
