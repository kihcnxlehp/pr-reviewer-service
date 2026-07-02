.PHONY: build run docker-up docker-down test

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down -v

test:
	go test ./...