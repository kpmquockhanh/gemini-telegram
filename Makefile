.PHONY: dev build build-armv7 stop logs clean

# Start development environment
up:
	docker-compose -f docker-compose.dev.yml up -d

# Start with logs attached
start:
	docker-compose -f docker-compose.dev.yml up

# Stop development environment
stop:
	docker-compose -f docker-compose.dev.yml stop

# Remove containers
down:
	docker-compose -f docker-compose.dev.yml down -v

# View logs
logs:
	docker-compose -f docker-compose.dev.yml logs -f

# Rebuild containers
build:
	docker-compose -f docker-compose.dev.yml build --no-cache

# Clean up everything
clean:
	docker-compose -f docker-compose.dev.yml down -v
	docker-compose -f docker-compose.dev.yml rm -f
	docker system prune -f

# Run API only
api:
	docker-compose -f docker-compose.dev.yml up -d api

# Run dashboard only
dashboard:
	docker-compose -f docker-compose.dev.yml up -d dashboard

# Development commands
dev:
	docker-compose -f docker-compose.dev.yml up api dashboard

# Restart API with fresh build
restart-api:
	docker-compose -f docker-compose.dev.yml up -d --build api

# Restart dashboard
restart-dashboard:
	docker-compose -f docker-compose.dev.yml up -d --build dashboard

# Enter API container shell
shell-api:
	docker-compose -f docker-compose.dev.yml exec api sh

# Enter dashboard container shell
shell-dashboard:
	docker-compose -f docker-compose.dev.yml exec dashboard sh

# Build and run binary
exec:
	go build -o bin/gemini-telegram-bot .
	./bin/gemini-telegram-bot

# Build binary for ARMv7 (e.g., Raspberry Pi)
build-armv7:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-s -w" -o bin/bot-armv7 .

# Run tests in container
test:
	docker-compose -f docker-compose.dev.yml exec api go test ./...
