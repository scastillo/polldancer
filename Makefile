.PHONY: all polldancer test clean help daemon run tidy run-webhook-server stop-webhook-server build-docker-image run-docker-container

all: help

polldancer: tidy
	@echo "Building polldancer..."
	go build -o polldancer ./cmd/polldancer/

test: run-webhook-server tidy
	@echo "Running tests..."
	go test -v -race ./...
	@$(MAKE) stop-webhook-server

tidy:
	@echo "Tidying and verifying module dependencies..."
	go mod tidy
	go mod verify

clean:
	@echo "Cleaning..."
	go clean
	rm -f polldancer

help:
	@echo "make - compile the application"
	@echo "make test - run tests"
	@echo "make clean - remove binary file and clean Go cache"
	@echo "make daemon - run application as a daemon"
	@echo "make run - run application"
	@echo "make tidy - tidy and verify module dependencies"
	@echo "make run-webhook-server - run webhook server"
	@echo "make stop-webhook-server - stop webhook server"
	@echo "make build-docker-image - build Docker image"
	@echo "make run-docker-container - run Docker container"

daemon: polldancer
	@echo "Running polldancer as a daemon..."
	nohup ./polldancer &

run: run-webhook-server polldancer
	@echo "Running polldancer..."
	@./polldancer & \
	echo $$! > polldancer.pid; \
	trap '$(MAKE) stop-webhook-server; kill `cat polldancer.pid` 2>/dev/null || true' SIGINT; \
	wait

run-webhook-server:
	@echo "Running webhook server..."
	python server.py & echo $$! > server.pid

stop-webhook-server:
	@echo "Stopping webhook server..."
	@-kill `cat server.pid` 2>/dev/null || true && rm -f server.pid

build-docker-image:
	@echo "Building Docker image..."
	docker build -t polldancer:latest .

run-docker-container:
	@echo "Running Docker container..."
	docker run -d --name polldancer-container polldancer:latest
