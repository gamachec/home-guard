.PHONY: build test dev-up dev-down install uninstall

build:
	GOOS=windows GOARCH=amd64 go build -ldflags "-H windowsgui" -o dist/home-guard.exe ./cmd/agent

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
