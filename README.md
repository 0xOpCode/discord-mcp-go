<div align="center">
  <h1>⚡ discord-mcp-go</h1>
  <p><strong>Ultra-fast, lightweight Discord Model Context Protocol (MCP) server written in Go.</strong></p>
  <p>
    <a href="https://github.com/0xOpCode/discord-mcp-go/blob/main/LICENSE"><img alt="License: MIT" src="https://img.shields.io/badge/License-MIT-blue.svg" /></a>
    <a href="https://golang.org"><img alt="Go Version" src="https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white" /></a>
    <a href="https://modelcontextprotocol.io"><img alt="MCP Protocol" src="https://img.shields.io/badge/MCP-Compatible-purple" /></a>
    <a href="https://discord.com"><img alt="Discord API" src="https://img.shields.io/badge/Discord-API%20v10-5865F2?logo=discord&logoColor=white" /></a>
  </p>
</div>

---

## ⚡ Overview

`discord-mcp-go` is a high-performance Discord MCP server designed for AI assistants and autonomous workflows (Claude Desktop, Cursor, n8n, Antigravity, OpenClaw). It wraps the Discord API into 75+ granular MCP tools, enabling full server administration, moderation, channel operations, and automated messaging.

### Why Go over Java Spring Boot?

| Metric | Original Java Server | discord-mcp-go | Improvement |
| :--- | :--- | :--- | :--- |
| **Startup Latency** | ~8 - 15 seconds | **< 15 milliseconds** | **~600x faster boot** |
| **RAM Footprint** | ~250MB - 400MB | **< 20 MB** | **93% memory savings** |
| **Multi-Server Control** | Static instance configuration | **Dynamic guild routing** | All bot servers managed via single token |
| **Transport Support** | stdio or Spring HTTP | **Streamable HTTP SSE + stdio** | Native dual transport |
| **Deployment Footprint** | Heavy JRE (~380MB image) | **Static Alpine container (~18MB)** | 20x smaller image |

---

## 🚀 Quickstart

### 1. Run with Docker Compose (Recommended)

Create a `.env` file:
```env
DISCORD_TOKEN=your_bot_token_here
DISCORD_GUILD_ID=optional_default_guild_id
PORT=8085
TRANSPORT=sse
```

Start the container:
```bash
docker compose up -d
```

Endpoint will be available at: `http://localhost:8085/sse` (and `http://localhost:8085/mcp`).

### 2. Run from Source

```bash
# Clone repository
git clone https://github.com/0xOpCode/discord-mcp-go.git
cd discord-mcp-go

# Run in HTTP SSE mode (default, port 8085)
DISCORD_TOKEN="your_bot_token" go run ./cmd/server --transport=sse --port=8085

# Run in STDIO mode
DISCORD_TOKEN="your_bot_token" go run ./cmd/server --transport=stdio
```

---

## 🔗 AI Client Setup

### Claude Desktop (`claude_desktop_config.json`)

#### Option A: HTTP SSE (Singleton mode, recommended)
Add under Connectors UI or via reverse proxy URL:
```json
{
  "mcpServers": {
    "discord": {
      "url": "http://localhost:8085/sse"
    }
  }
}
```

#### Option B: Stdio mode
```json
{
  "mcpServers": {
    "discord": {
      "command": "/absolute/path/to/discord-mcp-go",
      "args": ["-transport=stdio"],
      "env": {
        "DISCORD_TOKEN": "YOUR_DISCORD_BOT_TOKEN",
        "DISCORD_GUILD_ID": "OPTIONAL_DEFAULT_GUILD_ID"
      }
    }
  }
}
```

### Cursor (`~/.cursor/mcp.json`)
```json
{
  "mcpServers": {
    "discord": {
      "url": "http://localhost:8085/sse"
    }
  }
}
```

### n8n MCP Client Node
1. Add an **MCP Client** node to your n8n workflow.
2. Select **HTTP / SSE** transport.
3. Server URL: `http://localhost:8085/sse` (or container service name if internal).

---

## 🛠️ Complete Tool Directory (75+ Tools)

### Multi-Server & Discovery
- `list_servers`: List all Discord guilds joined by the bot with IDs and admin flags.
- `get_server_info`: Detailed guild metadata, counts, owner, and settings.
- `check_bot_permissions`: Comprehensive audit of bot permissions inside a guild.

### Message Management
- `send_message`: Post text messages to specific channels.
- `edit_message`: Edit bot-authored messages.
- `delete_message`: Remove messages from channels.
- `read_messages`: Paginated retrieval of channel history (cursor support: before, after, around).
- `add_reaction`: Add reactions using unicode or custom emojis.
- `remove_reaction`: Remove bot reactions from messages.

### Users & Direct Messages
- `get_user_id_by_name`: Resolve usernames or server nicknames to user IDs.
- `send_private_message`: Open DM channels and message users.
- `edit_private_message`: Edit sent direct messages.
- `delete_private_message`: Delete sent direct messages.
- `read_private_messages`: Retrieve direct message history.

### Channel Management
- `create_text_channel`: Create text channels with topics, category, slowmode, and NSFW settings.
- `edit_text_channel`: Update channel parameters.
- `delete_channel`: Delete text or voice channels.
- `find_channel`: Search channels by name substring.
- `list_channels`: List all server channels with IDs and types.
- `get_channel_info`: Fetch detailed channel properties.
- `move_channel`: Reposition channels and update category parentage.

### Category Management
- `create_category`: Create channel organizational categories.
- `edit_category`: Rename and reposition categories.
- `delete_category`: Delete channel categories.
- `find_category`: Find category by name.
- `list_channels_in_category`: List channels belonging to a category.

### Roles & Permissions
- `list_roles`: List all server roles, colors, and positions.
- `create_role`: Create roles with color, hoist, and permissions bitmask.
- `edit_role`: Update role metadata and permissions.
- `delete_role`: Remove roles from guilds.
- `assign_role`: Assign roles to server members.
- `remove_role`: Revoke roles from server members.
- `list_channel_permission_overwrites`: Audit channel permissions per role and member.
- `upsert_role_channel_permissions`: Set allow/deny bitmasks for roles.
- `upsert_member_channel_permissions`: Set allow/deny bitmasks for specific members.
- `delete_channel_permission_overwrite`: Reset permission overrides.

### Moderation
- `kick_member`: Kick members with audit log reasons.
- `ban_member`: Ban users with configurable message deletion windows.
- `unban_member`: Revoke bans.
- `timeout_member`: Apply communication timeouts (mutes) with minute-precision.
- `remove_timeout`: Lift communication timeouts immediately.
- `set_nickname`: Modify member nicknames.
- `get_bans`: List banned users and recorded reasons.

### Voice & Stage Channels
- `create_voice_channel`: Create voice channels with bitrate and user limits.
- `create_stage_channel`: Create stage channels for audio events.
- `edit_voice_channel`: Update voice channel parameters.
- `move_member`: Move active voice users between rooms.
- `disconnect_member`: Force-disconnect members from voice.
- `modify_voice_state`: Server mute or deafen members.

### Webhooks
- `create_webhook`: Create webhooks on target channels.
- `delete_webhook`: Remove webhooks.
- `list_webhooks`: View webhooks on a channel.
- `send_webhook_message`: Post messages with custom usernames via webhooks.

### Scheduled Events
- `create_guild_scheduled_event`: Schedule stage, voice, or external events with timestamps.
- `edit_guild_scheduled_event`: Update event details and lifecycle statuses.
- `delete_guild_scheduled_event`: Cancel and remove scheduled events.
- `list_guild_scheduled_events`: List upcoming server events.
- `get_guild_scheduled_event_users`: List subscribers interested in an event.

### Invites
- `create_invite`: Generate instant invites with max uses, age, and temporary status.
- `list_invites`: Audit active guild invites.
- `delete_invite`: Revoke invite codes.
- `get_invite_details`: Fetch invite destination and member counts.

### Forums
- `create_forum_channel`: Create forum discussion channels with tags and layout guidelines.
- `edit_forum_channel`: Update forum settings.
- `list_forum_channels`: List all server forum channels.
- `get_forum_channel_info`: Inspect forum tags and configuration.
- `list_forum_tags`: List tag names and moderation requirements.
- `create_forum_post`: Start threads with opening posts in forum channels.
- `list_forum_posts`: Fetch active discussion threads.
- `modify_forum_post`: Lock, archive, or rename forum threads.

### Custom Emojis
- `list_emojis`: List guild custom emojis.
- `get_emoji_details`: Inspect emoji author and animation flags.
- `create_emoji`: Upload new custom emojis using base64 image strings.
- `edit_emoji`: Rename custom emojis.
- `delete_emoji`: Remove custom emojis.

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).
