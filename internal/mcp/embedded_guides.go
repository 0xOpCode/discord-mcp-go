package mcp

const GuideOverview = `# Discord MCP Go: System Architecture & Agent Guide

## Overview
discord-mcp-go provides a Discord Model Context Protocol (MCP) server for AI assistants and autonomous agent platforms. It delivers 113 tools, native resource templates, an embedded persistent storage engine, and human-in-the-loop collaboration.

## Core Architecture
- **Dual Transport**: Supports Streamable HTTP Server-Sent Events (SSE) on /sse and standard input/output (stdio).
- **Tool Suite**: 113 tools across 18 operational domains.
- **Resource Templates**: Live Discord views under 'discord://...' schemas.
- **Persistent Engine**: JSON-backed local storage for custom agent memory and configuration.
- **Collaboration**: Direct user engagement via Discord DMs using 'ask_user'.

## Core Principles for Autonomous Agents
1. **Discover Before Mutating**: Read server overview resources before creating channels or roles.
2. **Batch Dependent Operations**: Group multiple actions into 'run_pipeline' to eliminate intermediate network latency.
3. **Gate Destructive Operations**: Request user approval through 'ask_user' before kicking members, purging messages, or deleting channels.
4. **Retain State in Memory**: Save runbooks, server notes, and user preferences to 'discord://memory/{key}'.
`

const GuideTools = `# Discord MCP Go: Tool Directory & Selection Rules

## Operational Domains (113 Tools)

### 1. Workflow Automation
- 'run_pipeline': Executes ordered steps with variable interpolation ('{{step1.id}}'). Use whenever you perform 2 or more related mutations.

### 2. Human-in-the-Loop Collaboration
- 'ask_user': Sends questions or approval requests to the human operator via Discord DM and blocks until a reply arrives. Supports screenshot and attachment retrieval.

### 3. Server Administration & Audit
- 'list_servers': Discovers all guilds accessible to the bot.
- 'get_server_info': Retrieves member counts, features, and owner details.
- 'check_bot_permissions': Audits bot permission bitmasks in a guild.
- 'get_audit_logs': Inspects administrative audit events.
- 'set_bot_activity': Sets presence activity (playing, listening, watching).

### 4. Messaging & Reactions
- 'send_message': Sends text, replies ('replyToMessageId'), or rich embeds ('embedsJson').
- 'read_messages': Retrieves paginated channel history.
- 'get_message': Fetches a message by channel ID and message ID.
- 'edit_message' / 'delete_message': Updates or removes bot messages.
- 'purge_messages': Bulk deletes up to 100 messages.
- 'add_reaction' / 'remove_reaction': Manages reaction emojis.
- 'pin_message' / 'unpin_message' / 'list_pinned_messages': Manages pinned channel messages.
- 'crosspost_message': Publishes announcement messages to following channels.
- 'search_guild_messages': Searches messages across server channels.
- 'send_file': Uploads local files within configured directory boundaries.

### 5. Channels & Categories
- 'create_text_channel' / 'edit_text_channel' / 'delete_channel': Manages text channels.
- 'create_voice_channel' / 'edit_voice_channel': Manages voice channels.
- 'create_stage_channel': Manages stage audio channels.
- 'create_category' / 'edit_category' / 'delete_category': Manages channel parent categories.
- 'list_channels' / 'get_channel_info' / 'find_channel': Channel inspection tools.
- 'move_channel': Updates channel sorting position and category parentage.
- 'create_thread' / 'create_thread_from_message' / 'list_threads': Thread lifecycle tools.

### 6. Roles & Permissions
- 'list_roles' / 'create_role' / 'edit_role' / 'delete_role': Server role definitions.
- 'assign_role' / 'remove_role': Toggles roles on members.
- 'upsert_role_channel_permissions' / 'upsert_member_channel_permissions': Channel overwrites.
- 'delete_channel_permission_overwrite': Clears permission overwrites.

### 7. Moderation & Safety
- 'kick_member' / 'ban_member' / 'unban_member' / 'get_bans': Member access control.
- 'timeout_member' / 'remove_timeout': Mutes users with expiration timestamps.
- 'disconnect_member' / 'move_member' / 'modify_voice_state': Voice moderation tools.
- 'prune_members' / 'estimate_prune': Inactive member cleanup tools.
- 'list_automod_rules' / 'create_automod_rule': Discord AutoMod configuration.

### 8. Custom Persistent Memory
- 'save_custom_resource': Stores state documents under 'discord://memory/{key}'.
- 'get_custom_resource': Reads stored documents by key or URI.
- 'delete_custom_resource': Deletes custom memory records.
- 'list_custom_resources': Lists all custom memory keys.

### 9. Media, Webhooks & Interactive Features
- 'create_poll' / 'end_poll': Discord interactive polls.
- 'list_webhooks' / 'create_webhook' / 'delete_webhook' / 'send_webhook_message': Webhook management.
- 'list_emojis' / 'get_emoji_details' / 'create_emoji' / 'edit_emoji' / 'delete_emoji': Emoji operations.
- 'list_stickers' / 'create_sticker' / 'delete_sticker': Sticker management.
- 'list_soundboard_sounds' / 'create_soundboard_sound' / 'delete_soundboard_sound': Audio clips.
- 'create_forum_channel' / 'list_forum_posts' / 'create_forum_post' / 'modify_forum_post': Forum channels.
- 'list_guild_scheduled_events' / 'create_guild_scheduled_event' / 'edit_guild_scheduled_event' / 'delete_guild_scheduled_event' / 'set_event_image': Calendar events.
`

const GuidePipelines = `# Discord MCP Go: Pipeline DSL & Recipes

## Pipeline Specification
The 'run_pipeline' tool executes a sequence of MCP actions as an atomic execution block. Each step runs in sequence, and outputs from earlier steps feed into subsequent steps via template interpolation.

### Schema
- 'steps' (array, required): Ordered array of step objects.
  - 'id' (string, required): Step identifier (e.g. 'step1', 'category', 'role').
  - 'tool' (string, required): Target MCP tool name (e.g. 'create_category').
  - 'arguments' (object, required): Tool input parameters.
- 'stopOnError' (boolean, optional, default: true): Halts execution if a step fails.

### Interpolation Syntax
Reference values from completed steps using double curly braces:
- '{{stepId.id}}': The Discord Snowflake ID extracted from the step output.
- '{{stepId.status}}': Execution status ('success' or 'failed').
- '{{stepId.output}}': Full text output produced by the step.

---

## Recipe 1: Community Onboarding Setup
Creates a parent category, child text channel, dedicated role, and welcome message in one round-trip:

` + "```json" + `
{
  "steps": [
    {
      "id": "cat",
      "tool": "create_category",
      "arguments": {
        "name": "Community Hub"
      }
    },
    {
      "id": "chan",
      "tool": "create_text_channel",
      "arguments": {
        "name": "welcome-and-rules",
        "parentId": "{{cat.id}}",
        "topic": "Official community guidelines and introductions"
      }
    },
    {
      "id": "msg",
      "tool": "send_message",
      "arguments": {
        "channelId": "{{chan.id}}",
        "content": "Welcome to the server! Read our pinned guidelines."
      }
    },
    {
      "id": "pin",
      "tool": "pin_message",
      "arguments": {
        "channelId": "{{chan.id}}",
        "messageId": "{{msg.id}}"
      }
    }
  ]
}
` + "```" + `

---

## Recipe 2: Event Announcement & Discussion Forum
Schedules a guild event, creates an accompanying announcement thread, and notifies members:

` + "```json" + `
{
  "steps": [
    {
      "id": "event",
      "tool": "create_guild_scheduled_event",
      "arguments": {
        "name": "AI Engineering Showcase",
        "entityType": "EXTERNAL",
        "location": "https://meet.example.com/ai-showcase",
        "startTime": "2026-10-15T18:00:00Z",
        "endTime": "2026-10-15T19:30:00Z",
        "description": "Demonstration of autonomous agent orchestration."
      }
    },
    {
      "id": "post",
      "tool": "send_message",
      "arguments": {
        "channelId": "1540398107504680963",
        "content": "📢 New Event Scheduled: **AI Engineering Showcase**! RSVP on the server event board."
      }
    }
  ]
}
` + "```" + `

---

## Recipe 3: Incident Containment
Locks down an affected channel, purges spam, and records the incident in persistent memory:

` + "```json" + `
{
  "steps": [
    {
      "id": "mute_role",
      "tool": "upsert_role_channel_permissions",
      "arguments": {
        "channelId": "1540398107504680963",
        "roleId": "1540398107504680961",
        "deny": ["SendMessages", "AddReactions"]
      }
    },
    {
      "id": "clean_spam",
      "tool": "purge_messages",
      "arguments": {
        "channelId": "1540398107504680963",
        "count": 50
      }
    },
    {
      "id": "log_event",
      "tool": "save_custom_resource",
      "arguments": {
        "key": "incident_lockdown_latest",
        "content": "Channel 1540398107504680963 locked down and purged 50 messages due to spam wave.",
        "description": "Security incident log"
      }
    }
  ]
}
` + "```" + `
`

const GuideBestPractices = `# Discord MCP Go: Agent Operating Protocols

## 1. Context Gathering Before Mutation
- **Use Native Resources First**: When you need server layout or channel lists, read 'discord://guilds/{guildId}/overview' instead of issuing multiple exploratory tool calls.
- **Inspect Role Hierarchies**: Read 'discord://guilds/{guildId}/roles' to determine existing permission sets before creating duplicate roles.
- **Review Channel History**: Read 'discord://channels/{channelId}/recent' to understand the topic context before posting messages.

## 2. Prioritize Pipelines Over Serial Tool Calls
- If a task involves two or more dependent Discord mutations (such as category creation, channel creation, and sending messages), combine them in 'run_pipeline'.
- Pipelines reduce network round-trips to one exchange and ensure structured rollback if a step fails.

## 3. Human Confirmation Policy ('ask_user')
Autonomous agents must gate irreversible or high-impact actions through 'ask_user':
- **Destructive Moderation**: Banning members, mass pruning members, or deleting roles.
- **Destructive Channel Deletion**: Deleting categories or channels with message history.
- **Mass Messaging**: Pinging '@everyone' or publishing announcements across multiple channels.

### Pattern:
1. Formulate the intended action plan.
2. Call 'ask_user' detailing the exact parameters and asking for confirmation.
3. Proceed only if the user approves.

## 4. Persistent Custom Memory
- Use 'save_custom_resource' to record persistent project settings, server guidelines, owner user ID ('owner_user_id'), or operational runbooks.
- All stored documents survive server restarts and map to 'discord://memory/{key}'.
`
