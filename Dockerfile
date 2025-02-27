# Use the official Golang image with Go 1.22 on Alpine
FROM golang:1.22-alpine

WORKDIR /app

# Copy the go.mod file
COPY go.mod ./

# Adjust the Go version in go.mod if necessary
RUN sed -i -E 's/^(go [0-9]+\.[0-9]+)\.[0-9]+/\1/' go.mod

# Download dependencies and generate go.sum
RUN go mod tidy

# Copy the entire source code
COPY . .

# Set the working directory to `api` where `main.go` is located
WORKDIR /app/api

# Default command (can be overridden in docker-compose)
CMD ["go", "run", "main.go"]
