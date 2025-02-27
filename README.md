
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
<img width="1470" alt="Screenshot 2025-02-27 at 8 28 09 PM" src="https://github.com/user-attachments/assets/a209d0b5-c7bd-428c-93da-3086a770ac35" />
<img width="1470" alt="Screenshot 2025-02-27 at 8 28 04 PM" src="https://github.com/user-attachments/assets/1716de47-6067-4e0b-a27e-fe3ee40c0eb0" />
<img width="1470" alt="Screenshot 2025-02-27 at 8 27 43 PM" src="https://github.com/user-attachments/assets/210aae72-43b7-4396-aa5a-12d6fa60945f" />
