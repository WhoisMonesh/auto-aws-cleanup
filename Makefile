.PHONY: build build-all clean test

VERSION ?= 0.1.0
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)

build:
	go build -ldflags "$(LDFLAGS)" -o auto-aws-cleanup .

build-all: build-mac build-linux build-windows

build-mac:
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o auto-aws-cleanup-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o auto-aws-cleanup-darwin-arm64 .

build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o auto-aws-cleanup-linux-amd64 .

build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o auto-aws-cleanup-windows-amd64.exe .

clean:
	rm -f auto-aws-cleanup auto-aws-cleanup-*

test:
	go test -v ./...
