.PHONY: build run serve cron migrate tidy fmt vet test lint

build:
	go build -o bin/go-starter .

run:
	go run . serve

serve:
	go run . serve

cron:
	go run . cron

migrate:
	go run . migrate

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./...

lint:
	golangci-lint run ./...
