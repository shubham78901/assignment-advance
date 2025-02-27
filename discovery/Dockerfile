FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY  . .

# Build the discovery service
RUN go build -o discovery discovery.go

# Create a minimal production image
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/discovery .

# Expose the default port
EXPOSE 8000

# Run the discovery service
CMD ["./discovery"]