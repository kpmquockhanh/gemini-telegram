#!/bin/bash

# Setup script for Raspberry Pi Zero 2 W
# This script downloads the correct binary and sets up the environment

set -e

echo "🍓 Setting up Gemini Telegram Bot on Raspberry Pi Zero 2 W..."

# Check architecture
ARCH=$(uname -m)
echo "Detected architecture: $ARCH"

if [ "$ARCH" == "aarch64" ]; then
    BINARY="bot-arm64"
    echo "✅ Using ARM64 binary (64-bit OS)"
elif [ "$ARCH" == "armv7l" ]; then
    BINARY="bot-armv7"
    echo "✅ Using ARMv7 binary (32-bit OS)"
else
    echo "⚠️  Unknown architecture: $ARCH"
    echo "Trying ARMv7 as fallback..."
    BINARY="bot-armv7"
fi

# Check if binary exists locally, if not download it
if [ ! -f "$BINARY" ]; then
    if [ -f "bot" ]; then
        echo "✅ Using existing 'bot' binary"
        BINARY="bot"
    else
        echo "❌ No binary found. Please download $BINARY from the releases page."
        echo "Or build it locally with: go build -o bot ."
        exit 1
    fi
fi

# Create .env if it doesn't exist
if [ ! -f ".env" ]; then
    echo ""
    echo "📝 Creating .env file..."
    cat > .env << 'EOF'
# Required: Get from @BotFather on Telegram
BOT_TOKEN=your_telegram_bot_token_here

# Required: Get from Google AI Studio (https://aistudio.google.com/app/apikey)
GEMINI_API_KEY=your_gemini_api_key_here

# Optional: Server port (default: 3000)
PORT=3000

# Optional: For Pi Zero 2 W, use 2 workers to save RAM
WORKER_POOL_SIZE=2

# Optional: SQLite database path
DATABASE_PATH=./data/prompts.db
EOF
    echo "⚠️  .env file created! You MUST edit it and add your tokens."
    echo "   nano .env"
    echo ""
    exit 1
fi

# Check if .env has been configured
if grep -q "your_telegram_bot_token_here" .env; then
    echo "❌ ERROR: .env file is not configured!"
    echo "   Please edit .env and add your real tokens:"
    echo "   nano .env"
    echo ""
    exit 1
fi

# Create data directory
mkdir -p data

# Check if binary is executable
if [ "$BINARY" != "bot" ] && [ -f "$BINARY" ]; then
    cp "$BINARY" bot
    chmod +x bot
fi

if [ ! -x "bot" ]; then
    chmod +x bot
fi

echo ""
echo "✅ Setup complete!"
echo "   Starting bot with WORKER_POOL_SIZE=2 (optimized for Pi Zero 2 W)..."
echo ""

# Run the bot
export WORKER_POOL_SIZE=2
./bot
