# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/discord-mcp-go ./cmd/server

# Final Stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bin/discord-mcp-go /app/discord-mcp-go

EXPOSE 8085

ENV PORT=8085
ENV TRANSPORT=sse

ENTRYPOINT ["/app/discord-mcp-go"]
