# Use a Go base image
FROM golang:1.16-alpine as builder

# Set the working directory
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download the Go module dependencies
RUN go mod download

# Copy the application source code
COPY . .

# Build the Go binary using the Makefile
RUN apk add --no-cache make && make polldancer

# Use a lightweight Alpine base image
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the Go binary from the builder stage
COPY --from=builder /app/polldancer .

# Expose the necessary port (if applicable)
# EXPOSE 8080

# Set the entrypoint command
CMD ["./polldancer"]
