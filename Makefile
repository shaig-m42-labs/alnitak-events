.PHONY: run test build

run:
	go run ./cmd/events

test:
	go test ./...

build:
	go build ./cmd/events
