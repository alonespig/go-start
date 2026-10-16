.PHONY: build run serve cron migrate tidy

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
