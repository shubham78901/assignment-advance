# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app

# Copy go.mod and go.sum; if go.sum is missing, create an empty file
COPY go.mod ./
RUN if [ ! -f go.sum ]; then echo "" > go.sum; fi
RUN go mod tidy

# Copy the rest of the application
COPY . .

# Build the application
RUN go build -o app .

# Final stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/app .
EXPOSE 8080
CMD ["./app"]
