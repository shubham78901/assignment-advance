FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /counter ./cmd/counter

# Create final lightweight image
FROM alpine:latest

WORKDIR /

# Copy binary from builder stage
COPY --from=builder /counter /counter

# Expose service and discovery ports
EXPOSE 8088
EXPOSE 8089/udp

# Set entrypoint
ENTRYPOINT ["/counter"]