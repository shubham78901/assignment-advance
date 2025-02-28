

# Distributed Counter System

A robust distributed counter system built with Go, featuring peer discovery, fault tolerance, and automatic synchronization between nodes.

## Overview

This distributed system implements a shared counter that remains consistent across multiple nodes. Key features include:

- Real-time counter synchronization across all nodes
- Automatic peer discovery through UDP broadcasting
- HTTP API for counter operations and peer management
- Health checking and automatic peer removal on failure
- Fault tolerance with automatic recovery
- Docker-based deployment for easy testing and scaling

## Architecture

The system consists of:

- **Counter Service**: Manages the counter state with thread-safe operations
- **Peers Manager**: Handles peer registration, discovery, and health checking
- **Discovery Service**: Enables automatic peer detection via UDP broadcasting
- **HTTP Handlers**: Provides REST API endpoints for system interaction

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Make (for running commands)
- curl (for testing)


## Testing

The system includes comprehensive tests to verify functionality:

# Run  Build containers and start

[make]


# Run all tests
[make test-full]



# Check node health
[make test-health]



# List peers for all nodes
[make test-all-peers]


# Test counter increment
[make test-increment]


# Check counter values across nodes
[make test-count]



# Test automatic peer discovery
[make test-discovery]



# Test peer removal and re-addition functionality
[make test-remove-existing-peer]


# Test peer registration
[make test-peers]


# Test counter synchronization
[make test-sync]
```

### Test Explanations

- **test-health**: Verifies all nodes are responsive and operational
- **test-count**: Shows current counter values on all nodes, confirming synchronization
- **test-discovery**: Demonstrates automatic discovery of new nodes
- **test-all-peers**: Lists all peer connections across nodes
- **test-peers**: Tests manual peer registration functionality
- **test-increment**: Increments the counter on one node and verifies propagation
- **test-sync**: Tests explicit counter synchronization across nodes
- **test-remove-existing-peer**: Tests system behavior when a node is removed and re-added

## API Reference

The system exposes the following HTTP endpoints:

- **GET /health**: Check node health
- **GET /count**: Get current counter value
- **POST /increment**: Increment counter and propagate to peers
- **POST /sync**: Update counter to match provided value if greater
- **GET /peers**: List known peers
- **POST /register**: Register a new peer
- **POST /remove-peer**: Remove a peer
- **GET /discovery**: Get discovery information











<img width="1470" alt="Screenshot 2025-02-26 at 12 33 30 PM" src="https://github.com/user-attachments/assets/b5d14ce9-e502-4d5b-8df2-45b5ba1cd80f" />
<img width="1470" alt="Screenshot 2025-02-27 at 8 27 43 PM" src="https://github.com/user-attachments/assets/210aae72-43b7-4396-aa5a-12d6fa60945f" />

<img width="1470" alt="Screenshot 2025-02-27 at 8 28 09 PM" src="https://github.com/user-attachments/assets/a209d0b5-c7bd-428c-93da-3086a770ac35" />
<img width="1470" alt="Screenshot 2025-02-27 at 8 28 04 PM" src="https://github.com/user-attachments/assets/1716de47-6067-4e0b-a27e-fe3ee40c0eb0" />
