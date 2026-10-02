<div align="center">
  <h1>⚡ discord-mcp-go</h1>
  <p><strong>Production Discord Model Context Protocol (MCP) server written in Go.</strong></p>
  <p>
    <a href="https://github.com/0xOpCode/discord-mcp-go/blob/main/LICENSE"><img alt="License: MIT" src="https://img.shields.io/badge/License-MIT-blue.svg" /></a>
    <a href="https://golang.org"><img alt="Go Version" src="https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white" /></a>
    <a href="https://modelcontextprotocol.io"><img alt="MCP Protocol" src="https://img.shields.io/badge/MCP-Compatible-purple" /></a>
    <a href="https://discord.com"><img alt="Discord API" src="https://img.shields.io/badge/Discord-API%20v10-5865F2?logo=discord&logoColor=white" /></a>
  </p>
</div>

---

## Overview

`discord-mcp-go` provides a production-grade Discord Model Context Protocol (MCP) server for AI assistants and autonomous agent platforms (Claude Desktop, Cursor, n8n, Antigravity, OpenClaw). It exposes 101 granular tools covering Discord server administration, moderation, channel operations, polls, stickers, soundboard clips, messaging, and multi-action block pipelines.

### Core Capabilities

- **Sub-15ms cold start**: Starts in under 15 milliseconds.
- **Low memory footprint**: Runs inside less than 20 MB RAM.
- **Dual transport architecture**: Native HTTP Server-Sent Events (SSE) and standard input/output (stdio).
- **Dynamic guild routing**: Manages all bot-joined Discord guilds through runtime `guildId` parameters and `DISCORD_GUILD_ID` defaults.
- **Blacklist protection**: Configurable guild blacklist (`DISCORD_BLACKLISTED_GUILDS`) to protect private servers against unauthorized agent actions.
- **Compound block execution**: Chained multi-action pipelines with variable interpolation via `run_pipeline`.
- **Minimal container footprint**: Multi-stage Alpine container image under 18 MB.

---

## 🚀 Quickstart

### 1. Run with Docker Compose (Recommended)

Create a `.env` file:
```env
DISCORD_TOKEN=your_bot_token_here
DISCORD_GUILD_ID=optional_default_guild_id
DISCORD_BLACKLISTED_GUILDS=optional_comma_separated_blacklisted_guild_ids
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

## 🛠️ Complete Tool Directory (101 Tools)

### Compound Workflows & Pipelines
- `run_pipeline`: Execute a sequence of MCP actions as connected blocks with variable reference interpolation (e.g. `{{step1.id}}`) in a single network round-trip.

### Multi-Server & Discovery
- `list_servers`: List all Discord guilds joined by the bot with IDs and admin flags.
- `get_server_info`: Detailed guild metadata, counts, owner, and settings.
- `check_bot_permissions`: Audit bot permissions inside a guild.
- `get_audit_logs`: Inspect audit log events and moderation actions.
- `set_bot_activity`: Set bot presence status (playing, watching, streaming, listening).

### Message Management
- `send_message`: Post text messages, message replies (`replyToMessageId`), or rich embeds (`embedsJson`) to specific channels.
- `get_message`: Fetch a specific message by channel ID and message ID.
- `send_file`: Upload files or attachments (images, PDFs, documents) to a text channel.
- `edit_message`: Edit bot-authored messages.
- `delete_message`: Remove messages from channels.
- `read_messages`: Paginated retrieval of channel history (cursor support: before, after, around).
- `purge_messages`: Bulk delete up to 100 recent messages from a channel.
- `add_reaction`: Add reactions using unicode or custom emojis.
- `remove_reaction`: Remove bot reactions from messages.

### Users & Direct Messages
- `get_user_id_by_name`: Resolve usernames or server nicknames to user IDs.
- `send_private_message`: Open DM channels and message users.
- `send_private_file`: Upload files or attachments directly to a user in DM.
- `edit_private_message`: Edit sent direct messages.
- `delete_private_message`: Delete sent direct messages.
- `read_private_messages`: Retrieve direct message history.

### Channel Management
- `create_text_channel`: Create text channels with topics, category, slowmode, and NSFW settings.
- `edit_text_channel`: Update channel parameters.
- `delete_channel`: Delete text or voice channels.
- `find_channel`: Search channels by name substring.
- `list_channels`: List all server channels with IDs and types.
- `get_channel_info`: Fetch channel properties.
- `move_channel`: Reposition channels and update category parentage.
- `create_thread`: Spawn public or private discussion threads inside text channels.
- `create_thread_from_message`: Start a public discussion thread directly on an existing message.
- `list_threads`: List active discussion threads within a text channel.

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

### Moderation & Member Pruning
- `kick_member`: Kick members with audit log reasons.
- `ban_member`: Ban users with configurable message deletion windows.
- `unban_member`: Revoke bans.
- `timeout_member`: Apply communication timeouts (mutes) with minute-precision.
- `remove_timeout`: Lift communication timeouts.
- `set_nickname`: Modify member nicknames.
- `get_bans`: List banned users and recorded reasons.
- `estimate_prune`: Estimate number of inactive members eligible for pruning.
- `prune_members`: Kick inactive members past specified days threshold.

### Auto-Moderation
- `list_automod_rules`: List active AutoMod keyword and spam filter rules.
- `create_automod_rule`: Configure custom keyword and spam blocking rules.

### Interactive Polls
- `create_poll`: Create native Discord polls with answers, duration hours, and multi-select.
- `end_poll`: Expire and finalize active polls.

### Voice & Stage Channels
- `create_voice_channel`: Create voice channels with bitrate and user limits.
- `create_stage_channel`: Create stage channels for audio events.
- `edit_voice_channel`: Update voice channel parameters.
- `move_member`: Move active voice users between rooms.
- `disconnect_member`: Disconnect members from voice.
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

### Soundboard Sounds
- `list_soundboard_sounds`: List custom soundboard audio clips.
- `create_soundboard_sound`: Upload custom soundboard audio clips with volume and emoji.
- `delete_soundboard_sound`: Remove soundboard clips.

### Custom Stickers
- `list_stickers`: List custom guild stickers.
- `create_sticker`: Upload custom stickers using base64 data or file paths.
- `delete_sticker`: Remove custom stickers.

### Role Connections & Linked Roles
- `get_role_connection`: Fetch user application role connection metadata.
- `update_role_connection`: Update application role connection platform and metadata.
- `get_role_connection_metadata`: Inspect application role connection metadata configuration.
- `update_role_connection_metadata`: Configure metadata records for verification rules.

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).
