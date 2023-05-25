
# Polldancer

Polldancer is a long poller service that gracefully dances to webhooks. We believe in humor and respect, so let's keep it light-hearted and inclusive.

## Introduction

Polldancer is a Go application that polls a remote service at regular intervals and forwards the received data to a webhook. It leverages the power of Go concurrency to efficiently handle multiple long-polling requests and gracefully dance to the rhythm of webhooks.

## Getting Started

### Prerequisites

- Go 1.16 or later
- Python 3.x (for running the webhook server)

### Configuration 
Polldancer can be configured using the following environment variables:

- `SLACK_TOKEN`: Slack API token for sending notifications to a Slack channel.
- `SLACK_CHANNEL`: Slack channel name to which notifications will be sent.
- `WEBHOOK_URL`: Webhook URL to forward the data to.
- `POLLING_URL`: URL to poll data from.
- `EXPECTED_MIME_TYPE`: Expected MIME type of the data from the polling URL.
- `POLLING_INTERVAL`: Interval between each poll in the format "5s" (default: 5 seconds).
- `RETRY_MAX_ATTEMPTS`: Maximum number of retry attempts before the circuit breaker trips (default: 3).

## Architecture

Polldancer follows the Onion architecture with Domain-Driven Design (DDD) principles to promote modularity, testability, and maintainability. The code is organized into different layers:

- `cmd`: Contains the main entry point of the application.
- `internal`: Contains the core logic of the application, including the polling, webhook, and Slack services.
- `pkg`: Contains reusable packages and interfaces.
- `mocks`: Contains the mock implementations of the interfaces for unit testing.
- `scripts`: Contains the helper scripts for building and running the application.
- `vendor`: Contains the application dependencies.

The recent improvements include the following updates:

- Added a circuit breaker implementation using the `gobreaker` library to prevent continuous retries during failures. The circuit breaker is applied to the polling and forwarding process.
- The circuit breaker allows a maximum number of retry attempts before temporarily stopping the execution.
- Logging is implemented using the Zap logger library for structured and configurable logging output.
- Each service is now initialized with a logger instance for consistent logging across the application.

### Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/yourusername/polldancer.git
   cd polldancer
   ```

2. Build the `polldancer` binary:

   ```bash
   make polldancer
   ```

### Docker

Polldancer can also be run as a Docker container. To build and run the Docker image, use the following commands:

```bash
# Build the Docker image
make build-image

# Run the Docker container
make run-container
```

The Docker image is built with the necessary environment variables configured through the `.env` file.

### Usage

Polldancer provides various options for running the application using the provided Makefile.

- **Running the Application**:

  To run Polldancer:

  ```bash
  make run
  ```

  This will start the Polldancer application, which will continuously poll a remote service and forward the data to the configured webhook.

- **Running Tests**:

  To run the tests:

  ```bash
  make test
  ```

  This will execute the tests and ensure the application behaves as expected.

- **Cleaning**:

  To clean up the project directory:

  ```bash
  make clean
  ```

  This will remove the `polldancer` binary and clean the Go cache.

- **Daemon Mode**:

  To run Polldancer as a daemon (in the background):

  ```bash
  make daemon
  ```

  This will start the Polldancer application as a daemon process, allowing it to run continuously in the background.

- **Tidying Dependencies**:

  To tidy and verify Go module dependencies:

  ```bash
  make tidy
  ```

  This will ensure that the Go module dependencies are properly managed and verified.

- **Running the Webhook Server**:

  To run the webhook server (required for Polldancer to forward data):

  ```bash
  make run-webhook-server
  ```

  This will start the webhook server, allowing it to receive and process forwarded data.

- **Stopping the Webhook Server**:

  To stop the webhook server:

  ```bash
  make stop-webhook-server
  ```

  This will gracefully stop the webhook server.

## Contributions and Code of Conduct

We welcome contributions from all developers. Please ensure that your contributions are respectful, inclusive, and adhere to our code of conduct. Let's create a supportive and positive environment for everyone involved.

## License

This project is licensed under the [MIT License](LICENSE).
```

Feel free to adjust and customize the README.md file according to your project's specific needs and branding.