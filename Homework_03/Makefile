.PHONY: build run test test-race lint cover docker-build docker-up docker-down clean

BINARY=bin/autodoc-api

build:
	go build -o $(BINARY) ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./... -v

test-race:
	go test ./... -race

cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run

docker-build:
	docker build -t autodoc-api:latest .

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down

clean:
	rm -rf bin/ coverage.out coverage.html
