package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/0xOpCode/discord-mcp-go/internal/discord"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type PipelineStep struct {
	ID        string                 `json:"id"`
	Tool      string                 `json:"tool"`
	Arguments map[string]interface{} `json:"arguments"`
}

type StepOutput struct {
	ID          string `json:"id"`
	Tool        string `json:"tool"`
	Success     bool   `json:"success"`
	Output      string `json:"output"`
	ExtractedID string `json:"extractedId,omitempty"`
	Error       string `json:"error,omitempty"`
}

var (
	rePlaceholder = regexp.MustCompile(`\{\{([a-zA-Z0-9_\-\.]+)\}\}|\$([a-zA-Z0-9_\-\.]+)`)
	reExplicitID  = regexp.MustCompile("(?i)(?:id[:=]\\s*[`\"']?([0-9]{17,20})[`\"']?)")
	reSnowflake   = regexp.MustCompile(`\b([0-9]{17,20})\b`)
)

func RegisterPipelineTools(s *server.MCPServer, client *discord.Client) {
	s.AddTool(
		mcp.NewTool("run_pipeline",
			mcp.WithDescription("Execute a sequence of MCP actions as connected blocks with variable reference interpolation"),
			mcp.WithString("steps", mcp.Required(), mcp.Description("JSON array of pipeline steps: [{id, tool, arguments}]")),
			mcp.WithBoolean("stopOnError", mcp.Description("Halt pipeline if a step returns error (default: true)")),
			mcp.WithString("guildId", mcp.Description("Default Discord Server ID applied to child steps")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			stepsRaw := req.Params.Arguments["steps"]
			steps, err := parsePipelineSteps(stepsRaw)
			if err != nil {
				return errorResult(err), nil
			}

			if len(steps) == 0 {
				return errorResult(fmt.Errorf("pipeline contains no steps")), nil
			}

			stopOnError := true
			if v, ok := req.Params.Arguments["stopOnError"].(bool); ok {
				stopOnError = v
			}

			defaultGuildID := getString(req.Params.Arguments, "guildId")
			contextMap := make(map[string]StepOutput)
			results := make([]StepOutput, 0, len(steps))

			for i, step := range steps {
				stepID := strings.TrimSpace(step.ID)
				if stepID == "" {
					stepID = fmt.Sprintf("step_%d", i+1)
				}

				if step.Tool == "run_pipeline" {
					return errorResult(fmt.Errorf("recursive run_pipeline invocation is disallowed at step '%s'", stepID)), nil
				}

				interpolatedArgs := interpolateArguments(step.Arguments, contextMap)
				if defaultGuildID != "" && interpolatedArgs["guildId"] == nil {
					interpolatedArgs["guildId"] = defaultGuildID
				}

				success, output, execErr := executeStep(ctx, s, step.Tool, interpolatedArgs)
				extractedID := extractID(output)

				stepRes := StepOutput{
					ID:          stepID,
					Tool:        step.Tool,
					Success:     success,
					Output:      output,
					ExtractedID: extractedID,
				}
				if execErr != nil {
					stepRes.Error = execErr.Error()
				}

				contextMap[stepID] = stepRes
				results = append(results, stepRes)

				if !success && stopOnError {
					break
				}
			}

			return successResult(formatPipelineResults(results)), nil
		},
	)
}

func parsePipelineSteps(arg interface{}) ([]PipelineStep, error) {
	switch v := arg.(type) {
	case string:
		var steps []PipelineStep
		if err := json.Unmarshal([]byte(v), &steps); err != nil {
			return nil, fmt.Errorf("invalid steps JSON: %w", err)
		}
		return steps, nil
	case []interface{}:
		raw, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		var steps []PipelineStep
		if err := json.Unmarshal(raw, &steps); err != nil {
			return nil, fmt.Errorf("invalid steps array: %w", err)
		}
		return steps, nil
	default:
		return nil, fmt.Errorf("steps parameter must be a JSON array or JSON string")
	}
}

func interpolateArguments(args map[string]interface{}, ctxMap map[string]StepOutput) map[string]interface{} {
	if args == nil {
		return make(map[string]interface{})
	}
	out := make(map[string]interface{}, len(args))
	for k, v := range args {
		out[k] = interpolateValue(v, ctxMap)
	}
	return out
}

func interpolateValue(val interface{}, ctxMap map[string]StepOutput) interface{} {
	switch v := val.(type) {
	case string:
		return interpolateString(v, ctxMap)
	case map[string]interface{}:
		return interpolateArguments(v, ctxMap)
	case []interface{}:
		arr := make([]interface{}, len(v))
		for i, item := range v {
			arr[i] = interpolateValue(item, ctxMap)
		}
		return arr
	default:
		return val
	}
}

func interpolateString(str string, ctxMap map[string]StepOutput) string {
	return rePlaceholder.ReplaceAllStringFunc(str, func(m string) string {
		var token string
		if strings.HasPrefix(m, "{{") && strings.HasSuffix(m, "}}") {
			token = m[2 : len(m)-2]
		} else if strings.HasPrefix(m, "$") {
			token = m[1:]
		}
		token = strings.TrimSpace(token)

		parts := strings.Split(token, ".")
		stepID := parts[0]
		field := "id"
		if len(parts) > 1 {
			field = strings.ToLower(parts[1])
		}

		res, ok := ctxMap[stepID]
		if !ok {
			return m
		}

		switch field {
		case "id":
			if res.ExtractedID != "" {
				return res.ExtractedID
			}
			return res.Output
		case "text", "raw", "output":
			return res.Output
		case "status":
			if res.Success {
				return "success"
			}
			return "failed"
		default:
			return m
		}
	})
}

func extractID(text string) string {
	if matches := reExplicitID.FindStringSubmatch(text); len(matches) > 1 {
		return matches[1]
	}
	if matches := reSnowflake.FindStringSubmatch(text); len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func executeStep(ctx context.Context, s *server.MCPServer, tool string, args map[string]interface{}) (bool, string, error) {
	reqMap := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      tool,
			"arguments": args,
		},
	}
	rawReq, err := json.Marshal(reqMap)
	if err != nil {
		return false, "", err
	}

	rpcResp := s.HandleMessage(ctx, rawReq)
	respBytes, err := json.Marshal(rpcResp)
	if err != nil {
		return false, "", err
	}

	var parsed struct {
		Result *mcp.CallToolResult `json:"result,omitempty"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}

	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		return false, "", err
	}

	if parsed.Error != nil {
		return false, "", fmt.Errorf("rpc error (%d): %s", parsed.Error.Code, parsed.Error.Message)
	}

	if parsed.Result == nil {
		return false, "", fmt.Errorf("no result returned from tool %s", tool)
	}

	var sb strings.Builder
	for _, c := range parsed.Result.Content {
		if rawMap, ok := c.(map[string]interface{}); ok {
			if txt, ok := rawMap["text"].(string); ok {
				sb.WriteString(txt)
			}
		} else if txtContent, ok := c.(mcp.TextContent); ok {
			sb.WriteString(txtContent.Text)
		} else {
			sb.WriteString(fmt.Sprintf("%v", c))
		}
	}

	output := strings.TrimSpace(sb.String())
	if parsed.Result.IsError {
		return false, output, fmt.Errorf("%s", output)
	}

	return true, output, nil
}

func formatPipelineResults(results []StepOutput) string {
	var sb strings.Builder
	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		}
	}

	sb.WriteString(fmt.Sprintf("### Pipeline Execution Summary (%d/%d steps succeeded):\n\n", successCount, len(results)))
	sb.WriteString("| Step | Tool | Status | Extracted ID | Output Preview |\n")
	sb.WriteString("|---|---|---|---|---|\n")

	for _, r := range results {
		status := "✅ Succeeded"
		if !r.Success {
			status = "❌ Failed"
		}

		extracted := "-"
		if r.ExtractedID != "" {
			extracted = fmt.Sprintf("`%s`", r.ExtractedID)
		}

		preview := r.Output
		if len(preview) > 60 {
			preview = preview[:57] + "..."
		}
		preview = strings.ReplaceAll(preview, "\n", " ")

		sb.WriteString(fmt.Sprintf("| `%s` | `%s` | %s | %s | %s |\n", r.ID, r.Tool, status, extracted, preview))
	}

	return sb.String()
}
