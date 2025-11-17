# Makefile for my-todo-learning

.PHONY: up down run build test fmt lint migrate clean

up:
	docker compose up -d

down:
	docker compose down -v

build:
	go build -o bin/my-todo-learning ./cmd/main.go

run:
	go run ./cmd/main.go

test:
	go test ./...

fmt:
	go fmt ./...

lint:
	go vet ./...

clean:
	rm -rf bin coverage.out coverage.html
