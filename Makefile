# Makefile for my-todo-learning

.PHONY: up down run build test test-verbose test-coverage test-coverage-func test-coverage-html fmt lint clean

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

test-verbose:
	go test -v ./...

test-coverage:
	go test -cover ./...

test-coverage-func:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

test-coverage-html:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

fmt:
	go fmt ./...

lint:
	go vet ./...

clean:
	rm -rf bin coverage.out coverage.html
