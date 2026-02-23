BINARY     = swarm-ctl
VERSION    = $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0")
BUILD_DATE = $(shell date -u +%Y-%m-%d)
LDFLAGS    = -ldflags="-s -w -X github.com/LyVanBong/swarm-ctl/cmd.Version=$(VERSION) -X github.com/LyVanBong/swarm-ctl/cmd.BuildDate=$(BUILD_DATE)"

.PHONY: build run clean install test lint release

## build: Build binary cho hệ thống hiện tại
build:
	@echo "🔨 Building $(BINARY) $(VERSION)..."
	@go build $(LDFLAGS) -o $(BINARY) .
	@echo "✅ Built: ./$(BINARY)"

## run: Build và chạy ngay
run: build
	@./$(BINARY)

## install: Cài vào /usr/local/bin
install: build
	@cp $(BINARY) /usr/local/bin/$(BINARY)
	@echo "✅ Installed to /usr/local/bin/$(BINARY)"

## clean: Xóa build artifacts
clean:
	@rm -f $(BINARY) $(BINARY)-*
	@echo "🧹 Cleaned"

## test: Chạy tests
test:
	@go test ./... -v

## lint: Chạy linter
lint:
	@golangci-lint run ./...

## release-linux: Build cho Linux (amd64 + arm64)
release-linux:
	@echo "🏗️  Building Linux binaries..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-linux-amd64 .
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY)-linux-arm64 .
	@echo "✅ Linux binaries built"

## release-all: Build cho tất cả platforms
release-all:
	@echo "🏗️  Building all platform binaries..."
	@GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-linux-amd64 .
	@GOOS=linux   GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY)-linux-arm64 .
	@GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-darwin-amd64 .
	@GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY)-darwin-arm64 .
	@ls -lah $(BINARY)-*
	@echo "✅ All platform binaries built"

## tidy: Update go.mod
tidy:
	@go mod tidy

## help: Show this help
help:
	@echo "swarm-ctl — Enterprise Docker Swarm Manager"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
