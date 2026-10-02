package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/0xOpCode/discord-mcp-go/internal/discord"
	"github.com/bwmarrin/discordgo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterWebhookTools(s *server.MCPServer, client *discord.Client) {
	// create_webhook
	s.AddTool(
		mcp.NewTool("create_webhook",
			mcp.WithDescription("Create a new webhook on a Discord channel"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Channel ID")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Webhook name")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			name := getString(req.Params.Arguments, "name")
			if channelID == "" || name == "" {
				return errorResult(fmt.Errorf("channelId and name are required")), nil
			}

			wh, err := client.Session.WebhookCreate(channelID, name, "")
			if err != nil {
				return errorResult(fmt.Errorf("failed to create webhook: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Webhook `%s` created. ID: `%s`, Token: `%s`", wh.Name, wh.ID, wh.Token)), nil
		},
	)

	// delete_webhook
	s.AddTool(
		mcp.NewTool("delete_webhook",
			mcp.WithDescription("Delete a webhook from Discord"),
			mcp.WithString("webhookId", mcp.Required(), mcp.Description("Webhook ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			webhookID := getString(req.Params.Arguments, "webhookId")
			if webhookID == "" {
				return errorResult(fmt.Errorf("webhookId is required")), nil
			}

			err := client.Session.WebhookDelete(webhookID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to delete webhook: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Webhook `%s` deleted successfully.", webhookID)), nil
		},
	)

	// list_webhooks
	s.AddTool(
		mcp.NewTool("list_webhooks",
			mcp.WithDescription("List all webhooks on a specific channel"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Channel ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			if channelID == "" {
				return errorResult(fmt.Errorf("channelId is required")), nil
			}

			webhooks, err := client.Session.ChannelWebhooks(channelID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to list webhooks: %w", err)), nil
			}

			if len(webhooks) == 0 {
				return successResult(fmt.Sprintf("No webhooks found on channel `%s`.", channelID)), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Webhooks on Channel `%s` (%d found):\n\n", channelID, len(webhooks)))
			sb.WriteString(fmt.Sprintf("| %-25s | %-20s | %-20s |\n", "Webhook Name", "Webhook ID", "Creator"))
			sb.WriteString("|---------------------------|----------------------|----------------------|\n")
			for _, wh := range webhooks {
				creator := "Unknown"
				if wh.User != nil {
					creator = wh.User.Username
				}
				sb.WriteString(fmt.Sprintf("| %-25s | %-20s | %-20s |\n", wh.Name, wh.ID, creator))
			}
			return successResult(sb.String()), nil
		},
	)

	// send_webhook_message
	s.AddTool(
		mcp.NewTool("send_webhook_message",
			mcp.WithDescription("Send a message through a Discord webhook"),
			mcp.WithString("webhookId", mcp.Required(), mcp.Description("Webhook ID")),
			mcp.WithString("webhookToken", mcp.Required(), mcp.Description("Webhook security token")),
			mcp.WithString("content", mcp.Required(), mcp.Description("Message content")),
			mcp.WithString("username", mcp.Description("Optional custom username override")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			webhookID := getString(req.Params.Arguments, "webhookId")
			webhookToken := getString(req.Params.Arguments, "webhookToken")
			content := getString(req.Params.Arguments, "content")
			if webhookID == "" || webhookToken == "" || content == "" {
				return errorResult(fmt.Errorf("webhookId, webhookToken, and content are required")), nil
			}

			params := &discordgo.WebhookParams{
				Content: content,
			}
			if u := getString(req.Params.Arguments, "username"); u != "" {
				params.Username = u
			}

			msg, err := client.Session.WebhookExecute(webhookID, webhookToken, true, params)
			if err != nil {
				return errorResult(fmt.Errorf("failed to send webhook message: %w", err)), nil
			}

			msgID := "executed"
			if msg != nil {
				msgID = msg.ID
			}

			return successResult(fmt.Sprintf("Webhook message sent successfully. ID: `%s`", msgID)), nil
		},
	)
}
