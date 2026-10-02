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

func RegisterAutoModTools(s *server.MCPServer, client *discord.Client) {
	// list_automod_rules
	s.AddTool(
		mcp.NewTool("list_automod_rules",
			mcp.WithDescription("List all configured Auto-Moderation rules in a Discord server"),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			rules, err := client.Session.AutoModerationRules(guildID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to fetch auto-moderation rules: %w", err)), nil
			}

			if len(rules) == 0 {
				return successResult(fmt.Sprintf("No auto-moderation rules configured in server `%s`.", guildID)), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Auto-Moderation Rules in Server `%s` (%d total):\n\n", guildID, len(rules)))
			sb.WriteString(fmt.Sprintf("| %-25s | %-20s | %-12s | %-8s |\n", "Rule Name", "Rule ID", "Trigger Type", "Enabled"))
			sb.WriteString("|---------------------------|----------------------|--------------|----------|\n")
			for _, r := range rules {
				name := r.Name
				if len(name) > 25 {
					name = name[:22] + "..."
				}
				enabled := false
				if r.Enabled != nil {
					enabled = *r.Enabled
				}
				sb.WriteString(fmt.Sprintf("| %-25s | %-20s | %-12d | %-8t |\n", name, r.ID, r.TriggerType, enabled))
			}
			return successResult(sb.String()), nil
		},
	)

	// create_automod_rule
	s.AddTool(
		mcp.NewTool("create_automod_rule",
			mcp.WithDescription("Create a keyword or spam filter Auto-Moderation rule in a server"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Rule name")),
			mcp.WithString("keywords", mcp.Description("Comma-separated list of blocked keywords/wildcards")),
			mcp.WithString("triggerType", mcp.Description("Trigger type: 'keyword', 'spam', 'mention_spam' (default: keyword)")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name := getString(req.Params.Arguments, "name")
			if name == "" {
				return errorResult(fmt.Errorf("name is required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			triggerTypeStr := strings.ToLower(getString(req.Params.Arguments, "triggerType"))
			var triggerType discordgo.AutoModerationRuleTriggerType = discordgo.AutoModerationEventTriggerKeyword
			if triggerTypeStr == "spam" {
				triggerType = discordgo.AutoModerationEventTriggerSpam
			} else if triggerTypeStr == "harmful_link" {
				triggerType = discordgo.AutoModerationEventTriggerHarmfulLink
			} else if triggerTypeStr == "keyword_preset" {
				triggerType = discordgo.AutoModerationEventTriggerKeywordPreset
			}

			var keywordsList []string
			if rawKeywords := getString(req.Params.Arguments, "keywords"); rawKeywords != "" {
				for _, kw := range strings.Split(rawKeywords, ",") {
					trimmed := strings.TrimSpace(kw)
					if trimmed != "" {
						keywordsList = append(keywordsList, trimmed)
					}
				}
			}

			enabled := true
			ruleData := &discordgo.AutoModerationRule{
				Name:        name,
				EventType:   discordgo.AutoModerationEventMessageSend,
				TriggerType: triggerType,
				Enabled:     &enabled,
				Actions: []discordgo.AutoModerationAction{
					{
						Type: discordgo.AutoModerationRuleActionBlockMessage,
					},
				},
			}

			if len(keywordsList) > 0 {
				ruleData.TriggerMetadata = &discordgo.AutoModerationTriggerMetadata{
					KeywordFilter: keywordsList,
				}
			}

			createdRule, err := client.Session.AutoModerationRuleCreate(guildID, ruleData)
			if err != nil {
				return errorResult(fmt.Errorf("failed to create auto-moderation rule: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Auto-Moderation rule `%s` created. ID: `%s`", createdRule.Name, createdRule.ID)), nil
		},
	)
}
