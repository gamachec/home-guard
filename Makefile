.PHONY: build test dev-up dev-down

build:
	GOOS=windows GOARCH=amd64 go build -ldflags "-H windowsgui" -o dist/home-guard.exe ./cmd/agent

test:
	go test ./...

dev-up:
	docker compose up -d

dev-down:
	docker compose down
