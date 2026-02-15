.PHONY: build build-updater test dev-up dev-down install uninstall

VERSION ?= dev

build:
	GOOS=windows GOARCH=amd64 go build -ldflags "-H windowsgui -X main.version=$(VERSION)" -o dist/home-guard.exe ./cmd/agent

build-updater:
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o dist/home-guard-updater.exe ./cmd/updater

test:
	go test ./...

install:
	dist/home-guard.exe install

uninstall:
	dist/home-guard.exe uninstall

dev-up:
	docker compose up -d

dev-down:
	docker compose down
