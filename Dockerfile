FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY  . .

# Build the node service
RUN go build -o node node.go

# Create a minimal production image
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/node .

# Expose the default port
EXPOSE 8088

# Run the node service
CMD ["./node"]