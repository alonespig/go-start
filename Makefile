.PHONY: build run serve cron migrate tidy fmt vet test lint rename-module

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

# 例：make rename-module MODULE=github.com/yourname/your-project
rename-module:
	@test -n "$(MODULE)" || (echo "用法: make rename-module MODULE=github.com/yourname/your-project" && exit 1)
	go run ./tools/rename-module $(MODULE)
