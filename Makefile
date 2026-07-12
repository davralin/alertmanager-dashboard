.PHONY: fmt test build compose-up compose-down

fmt:
	gofmt -w .

test:
	go test ./...

build:
	go build ./cmd/receiver
	go build ./cmd/dashboard

compose-up:
	docker compose up --build

compose-down:
	docker compose down --remove-orphans
