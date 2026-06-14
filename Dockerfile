# Stage 1: Build Vue frontend
FROM node:20-alpine AS vue-builder
WORKDIR /app/dashboard
COPY dashboard/package*.json ./
RUN npm ci
COPY dashboard/ .
RUN npm run build

# Stage 2: Build Go backend
FROM golang:1.24-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Copy Vue dist into build context
COPY --from=vue-builder /app/dashboard/dist ./dashboard/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bot .

# Stage 3: Runtime
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=go-builder /app/bot .
RUN mkdir -p /app/data
EXPOSE 3000
CMD ["./bot"]
