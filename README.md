# Gemini Telegram Bot (Go)

A Telegram bot that uses Google Gemini AI to generate images and videos from text prompts. Built with Go, using goroutines and channels for a native worker pool architecture.

## Features

- `/image <prompt>` — Generate images using Gemini Flash
- `/video <prompt>` — Generate videos using Veo 3.1
- Reply to a photo with `/image` or `/video` to use it as reference
- Real-time progress updates while generating
- Goroutine-based worker pool for handling concurrent requests
- SQLite persistence for default prompts per chat
- **Web Dashboard** — Manage prompts and templates via Vue 3 UI

## Architecture

```
main.go              HTTP server, Telegram webhook handler
api.go               REST API for dashboard
web.go               Embedded static files (Vue SPA) — production build
web_dev.go           Dev mode build tag (dashboard served by Vite)
config.go            Environment variables & configuration
bot.go               Telegram API wrapper (sendMessage, sendPhoto, etc.)
worker/
  pool.go            Goroutine worker pool with job queue
  job.go             Job types and result channels
ai/
  provider.go        Provider interface and registry
  providers/
    gemini/          Google Gemini provider (image + video)
    kling/           Kling AI provider (video)
storage/
  prompts.go         SQLite persistence for default prompts & templates
dashboard/           Vue 3 + Vite + Pinia frontend
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
# Build Vue frontend
cd dashboard && npm run build && cd ..

# Build Go binary
go build -o bot .
./bot
```

## Development

For development with hot reload, use the development Docker setup:

### Quick Start

```bash
# Start all services (API + Dashboard)
make dev
# Or:
./dev.sh dev
```

### Development Commands

```bash
# Start services in background
make up

# Start with logs
make start

# Stop services
make stop

# Rebuild containers
make build

# View logs
make logs

# Clean up
make clean

# Start only API
make api

# Start only Dashboard
make dashboard

# Enter API container shell
make shell-api

# Run tests
make test
```

### Development URLs

- **API**: http://localhost:3000
- **Dashboard**: http://localhost:8080

### How it works

- **API container** (`make api`): Runs Go backend with [Air](https://github.com/cosmtrek/air) for hot reload. Changes to `.go` files trigger automatic rebuild.
- **Dashboard container** (`make dashboard`): Runs Vite dev server with hot module replacement. Changes to Vue files are reflected immediately.
- **Volumes**: Source code is mounted into containers so changes are reflected without rebuilding.
- **SQLite**: Database file is persisted in `./data/prompts.db`.

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

## API Endpoints

### Telegram Webhook

| Method | Path | Description |
|---|---|---|
| POST | `/telegram` | Telegram webhook handler |

### Dashboard API

| Method | Path | Description |
|---|---|---|
| GET | `/` | Dashboard (Vue SPA) |
| GET | `/health` | Liveness probe |
| GET | `/api/stats` | Dashboard statistics |
| GET | `/api/prompts` | List prompts (paginated) |
| GET | `/api/prompts/{chatId}` | Get single prompt |
| PUT | `/api/prompts/{chatId}` | Update prompt |
| DELETE | `/api/prompts/{chatId}` | Delete prompt |
| GET | `/api/templates` | List templates |
| POST | `/api/templates` | Create template |
| PUT | `/api/templates/{id}` | Update template |
| DELETE | `/api/templates/{id}` | Delete template |

## Dashboard

Access the dashboard at `http://localhost:3000/` when running locally.

**Features:**
- Dashboard overview with statistics cards
- Prompts table with search, pagination, and filters
- Edit prompts with template application
- Copy prompt text to clipboard
- Template management (CRUD)
- Dark/light mode toggle
- Responsive sidebar navigation

## License

ISC
