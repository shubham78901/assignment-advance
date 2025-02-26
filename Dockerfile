# Use the official Golang image with Go 1.22 on Alpine
FROM golang:1.22-alpine

WORKDIR /app

# Copy the go.mod file
COPY go.mod ./

# (Optional) Adjust the go.mod version if necessary.
# This sed command strips off any patch version so "go 1.22.6" becomes "go 1.22"
RUN sed -i -E 's/^(go [0-9]+\.[0-9]+)\.[0-9]+/\1/' go.mod

# Download dependencies and generate go.sum
RUN go mod tidy

# Copy the rest of the source code
COPY . .

# Expose the ports used by your application
EXPOSE 8088 8089 8090

# Default command (can be overridden in docker-compose)
CMD ["go", "run", "main.go"]
