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

func RegisterPollTools(s *server.MCPServer, client *discord.Client) {
	// create_poll
	s.AddTool(
		mcp.NewTool("create_poll",
			mcp.WithDescription("Create an interactive poll in a Discord text channel"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Channel ID")),
			mcp.WithString("question", mcp.Required(), mcp.Description("Poll question text")),
			mcp.WithString("answers", mcp.Required(), mcp.Description("Comma-separated list of answer choices (2-10 options)")),
			mcp.WithNumber("durationHours", mcp.Description("Poll duration in hours: 1, 4, 8, 24, 72, or 168 (default: 24)")),
			mcp.WithBoolean("allowMultiselect", mcp.Description("Allow users to vote on multiple options")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			question := getString(req.Params.Arguments, "question")
			rawAnswers := getString(req.Params.Arguments, "answers")
			if channelID == "" || question == "" || rawAnswers == "" {
				return errorResult(fmt.Errorf("channelId, question, and answers are required")), nil
			}

			if err := client.CheckChannelAllowed(channelID); err != nil {
				return errorResult(err), nil
			}

			var answers []discordgo.PollAnswer
			for _, ans := range strings.Split(rawAnswers, ",") {
				trimmed := strings.TrimSpace(ans)
				if trimmed != "" {
					answers = append(answers, discordgo.PollAnswer{
						Media: &discordgo.PollMedia{
							Text: trimmed,
						},
					})
				}
			}

			if len(answers) < 2 || len(answers) > 10 {
				return errorResult(fmt.Errorf("polls require between 2 and 10 answer options (received %d)", len(answers))), nil
			}

			duration := getInt(req.Params.Arguments, "durationHours", 24)
			multiselect := getBool(req.Params.Arguments, "allowMultiselect", false)

			pollData := &discordgo.Poll{
				Question: discordgo.PollMedia{
					Text: question,
				},
				Answers:          answers,
				Duration:         duration,
				AllowMultiselect: multiselect,
				LayoutType:       discordgo.PollLayoutTypeDefault,
			}

			msg, err := client.Session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
				Poll: pollData,
			})
			if err != nil {
				return errorResult(fmt.Errorf("failed to create poll: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Poll created successfully in channel `%s`. Message ID: `%s`", channelID, msg.ID)), nil
		},
	)

	// end_poll
	s.AddTool(
		mcp.NewTool("end_poll",
			mcp.WithDescription("Immediately end an active poll and calculate final results"),
			mcp.WithString("channelId", mcp.Required(), mcp.Description("Channel ID containing the poll")),
			mcp.WithString("messageId", mcp.Required(), mcp.Description("Message ID of the poll")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID := getString(req.Params.Arguments, "channelId")
			messageID := getString(req.Params.Arguments, "messageId")
			if channelID == "" || messageID == "" {
				return errorResult(fmt.Errorf("channelId and messageId are required")), nil
			}

			if err := client.CheckChannelAllowed(channelID); err != nil {
				return errorResult(err), nil
			}

			endpoint := discordgo.EndpointChannelMessage(channelID, messageID) + "/poll/expire"
			_, err := client.Session.Request("POST", endpoint, nil)
			if err != nil {
				return errorResult(fmt.Errorf("failed to end poll: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Poll on message `%s` in channel `%s` ended.", messageID, channelID)), nil
		},
	)
}
