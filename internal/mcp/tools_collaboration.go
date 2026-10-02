package mcp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/0xOpCode/discord-mcp-go/internal/discord"
	"github.com/0xOpCode/discord-mcp-go/internal/storage"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func resolveTargetUserID(argID string, client *discord.Client, store *storage.Store) string {
	if trimmed := strings.TrimSpace(argID); trimmed != "" {
		return trimmed
	}
	if client != nil && strings.TrimSpace(client.OwnerUserID) != "" {
		return strings.TrimSpace(client.OwnerUserID)
	}
	if store != nil {
		if item, err := store.Get("owner_user_id"); err == nil && strings.TrimSpace(item.Content) != "" {
			return strings.TrimSpace(item.Content)
		}
	}
	return ""
}

func downloadAttachment(url, filename, destDir string) (string, error) {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}
	cleanName := filepath.Base(filename)
	if cleanName == "" || cleanName == "." || cleanName == "/" {
		cleanName = "attachment.bin"
	}
	destPath := filepath.Join(destDir, fmt.Sprintf("%d_%s", time.Now().UnixNano(), cleanName))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", err
	}
	return destPath, nil
}

func RegisterCollaborationTools(s *server.MCPServer, client *discord.Client, store *storage.Store) {
	s.AddTool(
		mcp.NewTool("ask_user",
			mcp.WithDescription("Ask the human user a question or confirmation via Discord DM and wait for their response"),
			mcp.WithString("question", mcp.Required(), mcp.Description("Question, prompt, or confirmation request to send to user")),
			mcp.WithString("userId", mcp.Description("Target user ID. If omitted, uses configured Owner ID or 'owner_user_id' stored in memory")),
			mcp.WithNumber("timeoutSeconds", mcp.Description("Timeout in seconds to wait for reply (5-600, default 120)")),
			mcp.WithBoolean("requireAttachment", mcp.Description("Wait until user provides an attachment or screenshot")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			targetUserID := resolveTargetUserID(getString(req.Params.Arguments, "userId"), client, store)
			if targetUserID == "" {
				return errorResult(fmt.Errorf("target user ID not provided: specify userId parameter, configure DISCORD_OWNER_USER_ID, or save custom resource 'owner_user_id'")), nil
			}

			if client == nil || client.Session == nil {
				return errorResult(fmt.Errorf("discord client not connected")), nil
			}

			question := getString(req.Params.Arguments, "question")
			if question == "" {
				return errorResult(fmt.Errorf("question parameter is required")), nil
			}

			timeoutSec := getInt(req.Params.Arguments, "timeoutSeconds", 120)
			if timeoutSec < 5 {
				timeoutSec = 5
			} else if timeoutSec > 600 {
				timeoutSec = 600
			}

			requireAttachment := getBool(req.Params.Arguments, "requireAttachment", false)

			dmChannel, err := client.Session.UserChannelCreate(targetUserID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to open DM channel with user %s: %w", targetUserID, err)), nil
			}

			sentMsg, err := client.Session.ChannelMessageSend(dmChannel.ID, question)
			if err != nil {
				return errorResult(fmt.Errorf("failed to send message to user %s: %w", targetUserID, err)), nil
			}

			timeoutTimer := time.NewTimer(time.Duration(timeoutSec) * time.Second)
			defer timeoutTimer.Stop()

			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()

			downloadDir := filepath.Join(os.TempDir(), "discord_attachments")

			for {
				select {
				case <-ctx.Done():
					return errorResult(ctx.Err()), nil
				case <-timeoutTimer.C:
					var sb strings.Builder
					sb.WriteString(fmt.Sprintf("### User Collaboration Timeout\n\nNo response received from user `%s` within %d seconds.\n\n", targetUserID, timeoutSec))
					sb.WriteString(fmt.Sprintf("**Question Sent**:\n> %s\n", strings.ReplaceAll(question, "\n", "\n> ")))
					return successResult(sb.String()), nil
				case <-ticker.C:
					messages, err := client.Session.ChannelMessages(dmChannel.ID, 10, "", sentMsg.ID, "")
					if err != nil {
						continue
					}

					for _, msg := range messages {
						if msg.Author == nil || msg.Author.ID != targetUserID {
							continue
						}
						if client.BotUser != nil && msg.Author.ID == client.BotUser.ID {
							continue
						}
						if requireAttachment && len(msg.Attachments) == 0 {
							continue
						}

						var downloadedPaths []string
						for _, att := range msg.Attachments {
							localPath, dErr := downloadAttachment(att.URL, att.Filename, downloadDir)
							if dErr == nil {
								downloadedPaths = append(downloadedPaths, fmt.Sprintf("- **%s** (Saved: `%s`, URL: %s)", att.Filename, localPath, att.URL))
							} else {
								downloadedPaths = append(downloadedPaths, fmt.Sprintf("- **%s** (URL: %s)", att.Filename, att.URL))
							}
						}

						var sb strings.Builder
						sb.WriteString("### User Response Received\n\n")
						sb.WriteString(fmt.Sprintf("- **Author**: %s (`%s`)\n", msg.Author.Username, msg.Author.ID))
						sb.WriteString(fmt.Sprintf("- **Timestamp**: %s\n", msg.Timestamp.Format("2006-01-02 15:04:05 UTC")))
						sb.WriteString(fmt.Sprintf("- **Message ID**: `%s`\n\n", msg.ID))
						sb.WriteString(fmt.Sprintf("#### Content\n%s\n", msg.Content))

						if len(downloadedPaths) > 0 {
							sb.WriteString(fmt.Sprintf("\n#### Attachments (%d)\n", len(downloadedPaths)))
							for _, p := range downloadedPaths {
								sb.WriteString(p + "\n")
							}
						}

						return successResult(sb.String()), nil
					}
				}
			}
		},
	)
}
