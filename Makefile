.PHONY: all build run clean test

APP_NAME=proxy

all: build

build:
	@echo "Building proxy server..."
	@go build -o bin/$(APP_NAME) cmd/proxy/main.go

run: build
	@echo "Running proxy server..."
	@./bin/$(APP_NAME)

clean:
	@echo "Cleaning up..."
	@go clean
	@rm -rf bin/ || if exist bin rmdir /s /q bin

test:
	@echo "Running tests..."
	@go test ./...
