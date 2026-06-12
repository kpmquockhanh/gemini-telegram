# Gemini Telegram Bot - Go Documentation

## Project Overview

A Telegram bot that uses Google Gemini AI to generate images and videos from text prompts. Built with Go using native goroutines and channels for concurrent processing.

---

## Table of Contents

1. [Architecture](#architecture)
2. [Getting Started](#getting-started)
3. [Environment Variables](#environment-variables)
4. [API Endpoints](#api-endpoints)
5. [Commands](#commands)
6. [Worker Pool](#worker-pool)
7. [Database Schema](#database-schema)
8. [Docker Deployment](#docker-deployment)
9. [Performance](#performance)
10. [Troubleshooting](#troubleshooting)

---

## Architecture

```
┌─────────────────┐
│   Telegram      │
│   Webhook       │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   HTTP Server   │
│   (main.go)     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Job Handler   │
│   (worker pool) │
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
┌───────┐ ┌───────┐
│Image  │ │Video  │
│Worker │ │Worker │
└───┬───┘ └───┬───┘
    │         │
    ▼         ▼
┌───────┐ ┌───────┐
│Gemini │ │Gemini │
│Flash  │ │Veo 3.1│
└───────┘ └───────┘
```

### File Structure

```
gemini-telegram-bot/
├── main.go              # Entry point, HTTP server, graceful shutdown
├── config.go            # Environment variables & configuration
├── bot.go               # Telegram API client wrapper
├── worker/
│   ├── pool.go          # Goroutine worker pool with job queue
│   └── job.go           # Job types and result channels
gemini/
│   ├── client.go        # Gemini SDK client initialization
│   ├── image.go         # Image generation logic
│   └── video.go         # Video generation + polling logic
storage/
│   └── prompts.go       # SQLite persistence for default prompts
├── go.mod
├── go.sum
├── Dockerfile
├── docker-compose.yml
└── README.md
```

---

## Getting Started

### Prerequisites

- Go 1.24 or later
- Docker (optional)
- Telegram Bot Token from [@BotFather](https://t.me/BotFather)
- Google Gemini API Key

### Local Development

1. **Clone the repository:**

```bash
git clone <repo-url>
cd gemini-telegram-bot
```

2. **Create a `.env` file:**

```bash
cp .env.example .env
```

Edit `.env` with your credentials:

```env
BOT_TOKEN=your_telegram_bot_token
GEMINI_API_KEY=your_google_gemini_api_key
PORT=3000
WORKER_POOL_SIZE=4
DATABASE_PATH=./data/prompts.db
```

3. **Install dependencies:**

```bash
go mod download
```

4. **Build the binary:**

```bash
go build -o bot .
```

5. **Run the server:**

```bash
./bot
```

### Setting the Webhook

After the server is running, set the Telegram webhook:

```bash
curl "https://api.telegram.org/bot<BOT_TOKEN>/setWebhook?url=https://<your-domain>/telegram"
```

**Note:** The webhook URL must be HTTPS and publicly accessible.

---

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BOT_TOKEN` | Yes | - | Telegram bot token from @BotFather |
| `GEMINI_API_KEY` | Yes | - | Google Gemini API key |
| `PORT` | No | `3000` | Server port |
| `WORKER_POOL_SIZE` | No | `4` | Number of worker goroutines |
| `DATABASE_PATH` | No | `./data/prompts.db` | SQLite database path |

---

## API Endpoints

### GET /

Health check endpoint. Returns "hello world".

**Response:**
```
hello world
```

### GET /health

Liveness probe endpoint. Returns HTTP 200 OK.

### POST /telegram

Telegram webhook endpoint. Processes incoming messages.

**Request Body:**
```json
{
  "message": {
    "message_id": 123,
    "chat": {
      "id": 123456789
    },
    "text": "/image a beautiful sunset"
  }
}
```

---

## Commands

### `/image <prompt>`

Generates an image using Gemini Flash.

**Examples:**
- `/image a beautiful sunset over mountains`
- `/image` (uses default prompt if set)
- Reply to a photo with `/image` to use it as reference

### `/video <prompt>`

Generates a video using Veo 3.1.

**Examples:**
- `/video a cat playing with a ball`
- `/video` (uses default prompt if set)
- Reply to a photo with `/video` to use it as reference

### `/set-image-prompt <prompt>`

Sets the default image prompt for the current chat.

**Example:**
- `/set-image-prompt a beautiful sunset`

### `/set-video-prompt <prompt>`

Sets the default video prompt for the current chat.

**Example:**
- `/set-video-prompt a cat playing`

---

## Worker Pool

### Design

The worker pool uses Go's native goroutines and channels for concurrent processing:

- **Jobs Channel:** Buffered channel for queuing work
- **Workers:** Fixed number of goroutines (configurable)
- **Result Channels:** Per-job result passing
- **Progress Callbacks:** Separate goroutine for progress updates

### Flow

1. HTTP handler receives request
2. Creates a `Job` with parameters
3. Enqueues job to the pool
4. Worker picks up job and processes it
5. Progress goroutine updates Telegram message every 10 seconds
6. Result is sent back via the job's `ResultChan`
7. HTTP handler sends the generated media

### Configuration

```go
pool := worker.NewPool(4, handleJob)
```

**Worker count:** Set via `WORKER_POOL_SIZE` environment variable.

---

## Database Schema

### SQLite Database

**Table:** `default_prompts`

```sql
CREATE TABLE IF NOT EXISTS default_prompts (
    chat_id INTEGER PRIMARY KEY,
    image_prompt TEXT,
    video_prompt TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Operations

**Get Prompts:**
```go
entry, err := store.GetPrompts(chatID)
```

**Set Image Prompt:**
```go
err := store.SetImagePrompt(chatID, prompt)
```

**Set Video Prompt:**
```go
err := store.SetVideoPrompt(chatID, prompt)
```

---

## Docker Deployment

### Build

```bash
docker-compose up --build
```

### Dockerfile

```dockerfile
# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies for CGO (SQLite)
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o bot .

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bot .

# Create data directory for SQLite
RUN mkdir -p /app/data

EXPOSE 3000

CMD ["./bot"]
```

### docker-compose.yml

```yaml
version: '3.8'

services:
  bot:
    build: .
    ports:
      - "3000:3000"
    env_file:
      - .env
    volumes:
      - ./data:/app/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:3000/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
```

---

## Performance

### Comparison: Node.js vs Go

| Metric | Node.js | Go | Improvement |
|--------|---------|-----|-------------|
| **Worker Memory** | ~4-8 MB/thread | ~2-4 KB/goroutine | **1000x smaller** |
| **Worker Startup** | ~100ms | ~1ms | **100x faster** |
| **Serialization** | `postMessage` + `TransferableArrayBuffer` | **Zero** (shared memory) | **None** |
| **Context Switching** | 1:1 OS threads | **M:N scheduler** | **More efficient** |
| **JSON Parsing** | `JSON.parse` | **Native** | **Faster** |
| **Signal Handling** | `process.on` | **Native `os/signal`** | **Better** |

### Why Go?

- **Goroutines:** Lightweight concurrent execution
- **Channels:** Safe communication between goroutines
- **Context:** Native cancellation and timeout support
- **Compiled:** Single binary, no runtime dependencies
- **Performance:** Efficient memory usage and fast execution

---

## Troubleshooting

### Common Issues

**1. Build fails with CGO errors:**

```bash
# Install build dependencies
apt-get install gcc musl-dev sqlite-dev
```

**2. SQLite database locked:**

SQLite supports limited concurrent writes. The code uses connection pooling and transactions to handle this.

**3. Webhook not receiving updates:**

- Ensure the webhook URL is HTTPS
- Check that the server is publicly accessible
- Verify the webhook is set correctly with Telegram

**4. Gemini API errors:**

- Check your `GEMINI_API_KEY` is valid
- Ensure the API has access to the models you're using
- Check rate limits

### Logs

The application uses structured logging with `log/slog`. Logs include:

- Job processing status
- Worker pool events
- API errors
- Progress updates

### Health Check

Use the `/health` endpoint to verify the server is running:

```bash
curl http://localhost:3000/health
```

---

## Development

### Adding New Commands

1. Add command handler in `main.go`:

```go
if strings.HasPrefix(text, "/newcommand") {
    // Handle command
    return
}
```

2. Implement the business logic in the appropriate package.

3. Update the README with the new command.

### Adding New Job Types

1. Add the job type in `worker/job.go`:

```go
const (
    JobTypeImage JobType = "image"
    JobTypeVideo JobType = "video"
    JobTypeNew   JobType = "new"
)
```

2. Add the handler in `main.go`:

```go
case worker.JobTypeNew:
    data, err := app.gemini.GenerateNew(ctx, job.Prompt)
```

3. Implement the generation logic in `gemini/`.

---

## Dependencies

### Go Modules

- `google.golang.org/genai` - Google GenAI SDK
- `github.com/mattn/go-sqlite3` - SQLite driver
- `github.com/joho/godotenv` - Environment variables

### Standard Library

- `net/http` - HTTP server
- `mime/multipart` - File uploads
- `encoding/json` - JSON parsing
- `log/slog` - Structured logging
- `context` - Context management
- `os/signal` - Signal handling

---

## License

ISC

---

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

---

## Support

For issues and questions:
- Open an issue on GitHub
- Contact the maintainer

---

**Last Updated:** 2026-06-12
**Version:** 2.0.0
**Language:** Go
