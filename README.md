# discord-mcp-go

Ultra-fast, lightweight Discord Model Context Protocol (MCP) server written in Go.

## Features

- **Blazing Fast**: Near-zero cold start (<15ms) and minimal RAM footprint (<25MB)
- **Multi-Server Management**: Single bot token controls all joined guilds/servers dynamically
- **Dual Transport**: Supports both Streamable HTTP SSE (`/mcp` and `/sse`) and standard `stdio`
- **75+ MCP Tools**: Comprehensive coverage of Discord REST & Gateway APIs (Messages, Channels, Roles, Moderation, Voice, Invites, Webhooks, Forums, Emojis)
- **Standalone Binary**: Zero JVM dependencies or runtime bloat

## Quickstart

### Prerequisites
- Go 1.22+
- Discord Bot Token

### Environment Variables
```bash
export DISCORD_TOKEN="your-bot-token"
export DISCORD_GUILD_ID="optional-default-guild-id"
export PORT="8085" # Optional, default 8085
```

### Run Server

```bash
# HTTP SSE mode (default)
go run ./cmd/server --transport=sse --port=8085

# STDIO mode
go run ./cmd/server --transport=stdio
```

## License
MIT
