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

func RegisterRoleTools(s *server.MCPServer, client *discord.Client) {
	// list_roles
	s.AddTool(
		mcp.NewTool("list_roles",
			mcp.WithDescription("List all roles in a Discord server"),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			roles, err := client.Session.GuildRoles(guildID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to fetch roles: %w", err)), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("### Server Roles (%d total):\n\n", len(roles)))
			sb.WriteString(fmt.Sprintf("| %-25s | %-20s | %-8s | %-8s |\n", "Role Name", "Role ID", "Color", "Position"))
			sb.WriteString("|---------------------------|----------------------|----------|----------|\n")
			for _, r := range roles {
				name := r.Name
				if len(name) > 25 {
					name = name[:22] + "..."
				}
				sb.WriteString(fmt.Sprintf("| %-25s | %-20s | #%06X | %-8d |\n", name, r.ID, r.Color, r.Position))
			}
			return successResult(sb.String()), nil
		},
	)

	// create_role
	s.AddTool(
		mcp.NewTool("create_role",
			mcp.WithDescription("Create a new role in a Discord server"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Role name")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
			mcp.WithNumber("color", mcp.Description("RGB color code (decimal)")),
			mcp.WithBoolean("hoist", mcp.Description("Whether the role is displayed separately")),
			mcp.WithBoolean("mentionable", mcp.Description("Whether anyone can mention this role")),
			mcp.WithNumber("permissions", mcp.Description("Permissions bitmask")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name := getString(req.Params.Arguments, "name")
			if name == "" {
				return errorResult(fmt.Errorf("role name is required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			params := &discordgo.RoleParams{
				Name: name,
			}
			if _, ok := req.Params.Arguments["color"]; ok {
				c := getInt(req.Params.Arguments, "color", 0)
				params.Color = &c
			}
			if _, ok := req.Params.Arguments["hoist"]; ok {
				h := getBool(req.Params.Arguments, "hoist", false)
				params.Hoist = &h
			}
			if _, ok := req.Params.Arguments["mentionable"]; ok {
				m := getBool(req.Params.Arguments, "mentionable", false)
				params.Mentionable = &m
			}
			if _, ok := req.Params.Arguments["permissions"]; ok {
				p := int64(getInt(req.Params.Arguments, "permissions", 0))
				params.Permissions = &p
			}

			role, err := client.Session.GuildRoleCreate(guildID, params)
			if err != nil {
				return errorResult(fmt.Errorf("failed to create role: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Role `%s` created successfully. ID: `%s`", role.Name, role.ID)), nil
		},
	)

	// edit_role
	s.AddTool(
		mcp.NewTool("edit_role",
			mcp.WithDescription("Edit an existing role in a Discord server"),
			mcp.WithString("roleId", mcp.Required(), mcp.Description("Role ID to edit")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
			mcp.WithString("name", mcp.Description("New role name")),
			mcp.WithNumber("color", mcp.Description("New color decimal")),
			mcp.WithBoolean("hoist", mcp.Description("Display separately")),
			mcp.WithBoolean("mentionable", mcp.Description("Mentionable")),
			mcp.WithNumber("permissions", mcp.Description("New permissions bitmask")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			roleID := getString(req.Params.Arguments, "roleId")
			if roleID == "" {
				return errorResult(fmt.Errorf("roleId is required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			params := &discordgo.RoleParams{}
			if name := getString(req.Params.Arguments, "name"); name != "" {
				params.Name = name
			}
			if _, ok := req.Params.Arguments["color"]; ok {
				c := getInt(req.Params.Arguments, "color", 0)
				params.Color = &c
			}
			if _, ok := req.Params.Arguments["hoist"]; ok {
				h := getBool(req.Params.Arguments, "hoist", false)
				params.Hoist = &h
			}
			if _, ok := req.Params.Arguments["mentionable"]; ok {
				m := getBool(req.Params.Arguments, "mentionable", false)
				params.Mentionable = &m
			}
			if _, ok := req.Params.Arguments["permissions"]; ok {
				p := int64(getInt(req.Params.Arguments, "permissions", 0))
				params.Permissions = &p
			}

			role, err := client.Session.GuildRoleEdit(guildID, roleID, params)
			if err != nil {
				return errorResult(fmt.Errorf("failed to edit role: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Role `%s` updated successfully. ID: `%s`", role.Name, role.ID)), nil
		},
	)

	// delete_role
	s.AddTool(
		mcp.NewTool("delete_role",
			mcp.WithDescription("Permanently delete a role from a Discord server"),
			mcp.WithString("roleId", mcp.Required(), mcp.Description("Role ID to delete")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			roleID := getString(req.Params.Arguments, "roleId")
			if roleID == "" {
				return errorResult(fmt.Errorf("roleId is required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			err = client.Session.GuildRoleDelete(guildID, roleID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to delete role: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Role `%s` deleted successfully.", roleID)), nil
		},
	)

	// assign_role
	s.AddTool(
		mcp.NewTool("assign_role",
			mcp.WithDescription("Assign a role to a server member"),
			mcp.WithString("userId", mcp.Required(), mcp.Description("Target user ID")),
			mcp.WithString("roleId", mcp.Required(), mcp.Description("Role ID to assign")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			userID := getString(req.Params.Arguments, "userId")
			roleID := getString(req.Params.Arguments, "roleId")
			if userID == "" || roleID == "" {
				return errorResult(fmt.Errorf("userId and roleId are required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			err = client.Session.GuildMemberRoleAdd(guildID, userID, roleID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to assign role: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Role `%s` assigned to user `%s`.", roleID, userID)), nil
		},
	)

	// remove_role
	s.AddTool(
		mcp.NewTool("remove_role",
			mcp.WithDescription("Remove a role from a server member"),
			mcp.WithString("userId", mcp.Required(), mcp.Description("Target user ID")),
			mcp.WithString("roleId", mcp.Required(), mcp.Description("Role ID to remove")),
			mcp.WithString("guildId", mcp.Description("Optional Discord Server ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			userID := getString(req.Params.Arguments, "userId")
			roleID := getString(req.Params.Arguments, "roleId")
			if userID == "" || roleID == "" {
				return errorResult(fmt.Errorf("userId and roleId are required")), nil
			}

			guildID, err := client.ResolveGuildID(getString(req.Params.Arguments, "guildId"))
			if err != nil {
				return errorResult(err), nil
			}

			err = client.Session.GuildMemberRoleRemove(guildID, userID, roleID)
			if err != nil {
				return errorResult(fmt.Errorf("failed to remove role: %w", err)), nil
			}

			return successResult(fmt.Sprintf("Role `%s` removed from user `%s`.", roleID, userID)), nil
		},
	)
}
