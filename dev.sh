#!/bin/sh
# Docker Development Helper Script
# Usage: ./dev.sh [up|start|stop|down|logs|build|clean|api|dashboard|dev]

COMPOSE_FILE="docker-compose.dev.yml"

case "$1" in
  up|start)
    docker-compose -f $COMPOSE_FILE up -d
    echo "Development environment started!"
    echo "API: http://localhost:3000"
    echo "Dashboard: http://localhost:8080"
    ;;
  
  stop)
    docker-compose -f $COMPOSE_FILE stop
    ;;
  
  down)
    docker-compose -f $COMPOSE_FILE down -v
    ;;
  
  logs)
    docker-compose -f $COMPOSE_FILE logs -f
    ;;
  
  build)
    docker-compose -f $COMPOSE_FILE build --no-cache
    ;;
  
  clean)
    docker-compose -f $COMPOSE_FILE down -v
    docker-compose -f $COMPOSE_FILE rm -f
    docker system prune -f
    ;;
  
  api)
    docker-compose -f $COMPOSE_FILE up -d api
    ;;
  
  dashboard)
    docker-compose -f $COMPOSE_FILE up -d dashboard
    ;;
  
  dev)
    docker-compose -f $COMPOSE_FILE up api dashboard
    ;;
  
  shell-api)
    docker-compose -f $COMPOSE_FILE exec api sh
    ;;
  
  shell-dashboard)
    docker-compose -f $COMPOSE_FILE exec dashboard sh
    ;;
  
  test)
    docker-compose -f $COMPOSE_FILE exec api go test ./...
    ;;
  
  *)
    echo "Docker Development Helper"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  up           Start development environment in background"
    echo "  start        Start development environment with logs"
    echo "  stop         Stop development environment"
    echo "  down         Stop and remove containers"
    echo "  logs         View logs"
    echo "  build        Rebuild containers"
    echo "  clean        Clean up everything"
    echo "  api          Start only API container"
    echo "  dashboard    Start only dashboard container"
    echo "  dev          Start API and dashboard with logs"
    echo "  shell-api    Enter API container shell"
    echo "  shell-dashboard  Enter dashboard container shell"
    echo "  test         Run tests in API container"
    echo ""
    echo "URLs:"
    echo "  API:       http://localhost:3000"
    echo "  Dashboard: http://localhost:8080"
    ;;
esac
