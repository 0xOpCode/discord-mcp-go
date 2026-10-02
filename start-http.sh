#!/bin/bash
set -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PIDFILE="$DIR/discord-mcp.pid"
LOGFILE="$DIR/discord-http.log"

export DISCORD_TOKEN="${DISCORD_TOKEN:-MTUxMzg5ODQ0MTc0NTYzMzUyMg.GCnFsF.N5BQRcRmFYY9Jgc545_59vHj_vPgxZS0v0hpd4}"
export DISCORD_GUILD_ID="${DISCORD_GUILD_ID:-1540398107504680961}"
export PORT="8085"
export TRANSPORT="sse"

if [ -f "$PIDFILE" ]; then
    OLD_PID=$(cat "$PIDFILE" 2>/dev/null || true)
    if [ -n "$OLD_PID" ] && kill -0 "$OLD_PID" 2>/dev/null; then
        kill "$OLD_PID" 2>/dev/null || true
        sleep 0.5
        kill -9 "$OLD_PID" 2>/dev/null || true
    fi
    rm -f "$PIDFILE"
fi

fuser -k 8085/tcp 2>/dev/null || true
pkill -9 -f "discord-mcp-go" 2>/dev/null || true

setsid /usr/local/bin/discord-mcp-go -transport=sse -port=8085 </dev/null > "$LOGFILE" 2>&1 &
NEW_PID=$!
disown -h $NEW_PID 2>/dev/null || true
echo "$NEW_PID" > "$PIDFILE"

if [ "$1" = "--daemon" ] || [ "$1" = "-d" ]; then
    exit 0
fi

for i in {1..20}; do
    if curl -s -f http://127.0.0.1:8085/sse >/dev/null 2>&1 || nc -z 127.0.0.1 8085 2>/dev/null; then
        echo "✅ Discord MCP Go Server is UP on http://127.0.0.1:8085/sse (PID: $NEW_PID)"
        exit 0
    fi
    sleep 0.5
done
