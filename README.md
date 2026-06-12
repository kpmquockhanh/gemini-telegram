# Gemini Telegram Bot (Go)

A Telegram bot that uses Google Gemini AI to generate images and videos from text prompts. Built with Go, using goroutines and channels for a native worker pool architecture.

## Features

- `/image <prompt>` — Generate images using Gemini Flash
- `/video <prompt>` — Generate videos using Veo 3.1
- Reply to a photo with `/image` or `/video` to use it as reference
- Real-time progress updates while generating
- Goroutine-based worker pool for handling concurrent requests
- SQLite persistence for default prompts per chat

## Architecture

```
main.go              HTTP server, Telegram webhook handler
config.go            Environment variables & configuration
bot.go               Telegram API wrapper (sendMessage, sendPhoto, etc.)
worker/
  pool.go            Goroutine worker pool with job queue
  job.go             Job types and result channels
gemini/
  client.go          Gemini SDK client initialization
  image.go           Image generation logic
  video.go           Video generation + polling logic
storage/
  prompts.go         SQLite persistence for default prompts
```

## Setup

### 1. Clone the repo and build:

```bash
git clone <repo-url>
cd gemini-telegram-bot
```

### 2. Create a `.env` file:

```
BOT_TOKEN=<your-telegram-bot-token>
GEMINI_API_KEY=<your-google-gemini-api-key>
PORT=3000
WORKER_POOL_SIZE=4
DATABASE_PATH=./data/prompts.db
```

### 3. Build and run with Docker:

```bash
docker-compose up --build
```

Or build locally:

```bash
go build -o bot .
./bot
```

### 4. Set the Telegram webhook:

Make sure the server is running, then call:

```bash
curl "https://api.telegram.org/bot<BOT_TOKEN>/setWebhook?url=https://<your-domain>/telegram"
```

Replace `<BOT_TOKEN>` with your actual token and `<your-domain>` with your deployed server URL. The endpoint must be HTTPS and publicly accessible (use ngrok for local testing).

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `BOT_TOKEN` | Yes | — | Telegram bot token from @BotFather |
| `GEMINI_API_KEY` | Yes | — | Google Gemini API key |
| `PORT` | No | `3000` | Server port |
| `WORKER_POOL_SIZE` | No | `4` | Number of worker goroutines |
| `DATABASE_PATH` | No | `./data/prompts.db` | SQLite database path |

## Why Go?

- **Goroutines**: ~2-4KB memory per worker vs ~4-8MB per Node.js thread
- **No serialization overhead**: Jobs share memory directly (no `postMessage`/`TransferableArrayBuffer`)
- **Native context cancellation**: Clean shutdown with `context.Context`
- **SQLite**: Better concurrency than JSON file for prompt storage

## API Endpoints

| Method | Path | Description |
|---|---|---|
| GET | `/` | Health check ("hello world") |
| GET | `/health` | Liveness probe |
| POST | `/telegram` | Telegram webhook handler |

## License

ISC
