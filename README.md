# gemini-telegram-bot

A Telegram bot that uses Google Gemini AI to generate images and videos from text prompts. Built with Express and a multi-threaded worker pool for concurrent generation tasks.

## Features

- `/image <prompt>` — Generate images using Gemini Flash
- `/video <prompt>` — Generate videos using Veo 3.1
- Reply to a photo with `/image` or `/video` to use it as reference
- Real-time progress updates while generating
- Worker pool architecture for handling concurrent requests

## Setup

1. Clone the repo and install dependencies:

```bash
git clone <repo-url>
cd gemini-telegram-bot
npm install
```

2. Create a `.env` file:

```
BOT_TOKEN=<your-telegram-bot-token>
GEMINI_API_KEY=<your-google-gemini-api-key>
PORT=3000
WORKER_POOL_SIZE=4
```

3. Set up a Telegram bot webhook pointing to `https://<your-domain>/telegram`.

4. Start the server:

```bash
npm start
```

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `BOT_TOKEN` | Yes | Telegram bot token from [@BotFather](https://t.me/BotFather) |
| `GEMINI_API_KEY` | Yes | Google Gemini API key |
| `PORT` | No | Server port (default: `3000`) |
| `WORKER_POOL_SIZE` | No | Number of worker threads (default: `4`) |

## Architecture

```
index.js              Express server, Telegram webhook handler
WorkerPool.js         Thread pool manager with job queue
workers/ai-worker.js  Worker thread running Gemini AI calls
google-ai.js          Gemini API wrapper (unused, kept for reference)
```

Requests are enqueued into a worker pool. Each worker runs in its own Node.js thread and communicates results back via `postMessage`. The pool auto-scales by replacing crashed workers.
