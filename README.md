
# Distributed Counter

A distributed counter service that maintains consistent counts across multiple nodes.

## Features

- Peer-to-peer architecture
- Automatic service discovery via UDP multicast
- HTTP API for counter operations
- Automatic peer health checking
- Counter value propagation to maintain consistency

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Docker (optional for containerized deployment)


## API Endpoints

- `GET /count` - Get the current counter value
- `POST /increment` - Increment the counter
- `POST /sync` - Sync counter with provided value
- `GET /peers` - List all known peers
- `POST /register` - Register a new peer
- `POST /remove-peer` - Remove a peer
- `GET /health` - Check node health
- `GET /discovery` - Get discovery information

## Architecture

The system is designed with the following components:

- **Config**: Application configuration and node identification
- **Counter**: Manages the counter value with thread-safe operations
- **Discovery**: Handles peer discovery via UDP broadcasts
- **Peers**: Manages peer connections and health checks
- **Handlers**: HTTP API handlers for all endpoints

## License

MIT
Run [make]



Run [make test-full]

<img width="1470" alt="Screenshot 2025-02-26 at 12 33 30 PM" src="https://github.com/user-attachments/assets/b5d14ce9-e502-4d5b-8df2-45b5ba1cd80f" />
<img width="1470" alt="Screenshot 2025-02-25 at 7 55 03 PM" src="https://github.com/user-attachments/assets/d10401f5-be0c-432b-b593-c27f0c4f7673" />
<img width="1470" alt="Screenshot 2025-02-26 at 12 54 53 PM" src="https://github.com/user-attachments/assets/f1f01a27-0940-454e-9ddc-dfb81eb8c496" />
