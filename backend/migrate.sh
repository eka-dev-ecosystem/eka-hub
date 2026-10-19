#!/bin/bash

# Migration helper script for Linux/Mac

set -e

ACTION=${1:-"help"}
NAME=${2:-""}
STEPS=${3:-1}
VERSION=${4:-0}

case "$ACTION" in
  up)
    echo "Running migrations..."
    go run cmd/migrate.go -action=up
    ;;
  down)
    echo "Rolling back $STEPS migration(s)..."
    go run cmd/migrate.go -action=down -steps=$STEPS
    ;;
  create)
    if [ -z "$NAME" ]; then
      echo "Error: Please specify migration name"
      echo "Usage: ./migrate.sh create migration_name"
      exit 1
    fi
    echo "Creating migration: $NAME"
    go run cmd/migrate.go -action=create -name=$NAME
    ;;
  version)
    echo "Checking migration version..."
    go run cmd/migrate.go -action=version
    ;;
  force)
    if [ "$VERSION" -eq 0 ]; then
      echo "Error: Please specify version"
      echo "Usage: ./migrate.sh force name version"
      exit 1
    fi
    echo "Forcing migration to version $VERSION..."
    go run cmd/migrate.go -action=force -version=$VERSION
    ;;
  drop)
    echo "WARNING: This will drop all tables!"
    read -p "Are you sure? Type 'yes' to confirm: " confirm
    if [ "$confirm" != "yes" ]; then
      echo "Aborted"
      exit 0
    fi
    go run cmd/migrate.go -action=drop
    ;;
  help|*)
    echo "Migration Helper Script"
    echo ""
    echo "Usage: ./migrate.sh [command] [args]"
    echo ""
    echo "Commands:"
    echo "  up                      - Run all pending migrations"
    echo "  down [steps]            - Rollback migrations (default: 1 step)"
    echo "  create <name>           - Create new migration files"
    echo "  version                 - Show current migration version"
    echo "  force <name> <version>  - Force migration version (use with caution)"
    echo "  drop                    - Drop all tables (dangerous!)"
    echo "  help                    - Show this help message"
    echo ""
    echo "Examples:"
    echo "  ./migrate.sh up"
    echo "  ./migrate.sh down 2"
    echo "  ./migrate.sh create add_users_table"
    echo "  ./migrate.sh version"
    echo "  ./migrate.sh force _ 1"
    echo ""
    ;;
esac

