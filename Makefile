.PHONY: help build run test clean docker-build docker-up docker-down install

help:
	@echo "Available commands:"
	@echo "  make install       - Install dependencies"
	@echo "  make build         - Build the application"
	@echo "  make run           - Run the application"
	@echo "  make test          - Run tests"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-up     - Start Docker containers"
	@echo "  make docker-down   - Stop Docker containers"

install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

build:
	@echo "Building API Gateway..."
	go build -o bin/gateway cmd/gateway/main.go

run:
	@echo "Starting API Gateway..."
	go run cmd/gateway/main.go

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f go.sum

docker-build:
	@echo "Building Docker image..."
	docker build -t api-gateway:latest -f deployments/Dockerfile .

docker-up:
	@echo "Starting Docker containers..."
	docker-compose -f deployments/docker-compose.yml up -d

docker-down:
	@echo "Stopping Docker containers..."
	docker-compose -f deployments/docker-compose.yml down

lint:
	@echo "Running linter..."
	golangci-lint run
