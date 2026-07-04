#!/bin/bash

# Script to easily run or build the Twistgram API Backend

show_help() {
    echo "Usage: ./run.sh [command]"
    echo ""
    echo "Commands:"
    echo "  start       Run the API server with live reloading (requires Air)"
    echo "  run         Run the API server normally (go run)"
    echo "  build       Build the API server binary"
    echo "  test        Run all unit tests"
    echo "  tidy        Tidy up go.mod dependencies"
    echo "  help        Show this help message"
}

case "$1" in
    start)
        echo "Starting Twistgram API with Air (live reload)..."
        if ! command -v air &> /dev/null; then
            echo "Air is not installed. Installing..."
            go install github.com/air-verse/air@latest
        fi
        air
        ;;
    run)
        echo "Running Twistgram API..."
        go run cmd/api/main.go
        ;;
    build)
        echo "Building Twistgram API binary..."
        go build -o twistgram-api cmd/api/main.go
        echo "Build complete: ./twistgram-api"
        ;;
    test)
        echo "Running Twistgram API tests..."
        go test ./... -v
        ;;
    tidy)
        echo "Tidying dependencies..."
        go mod tidy
        ;;
    *)
        show_help
        ;;
esac
