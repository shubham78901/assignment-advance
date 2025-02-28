
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
<img width="1470" alt="Screenshot 2025-02-27 at 8 27 43 PM" src="https://github.com/user-attachments/assets/210aae72-43b7-4396-aa5a-12d6fa60945f" />

<img width="1470" alt="Screenshot 2025-02-27 at 8 28 09 PM" src="https://github.com/user-attachments/assets/a209d0b5-c7bd-428c-93da-3086a770ac35" />
<img width="1470" alt="Screenshot 2025-02-27 at 8 28 04 PM" src="https://github.com/user-attachments/assets/1716de47-6067-4e0b-a27e-fe3ee40c0eb0" />












base) shubhamgautam@shubhams-MacBook-Air assignment-advance % make test-remove-existing-peer
🔍 Testing removal of an existing peer...
1. Checking current peers on all nodes:
Node1 peers:
[
  "172.18.0.3:8090",
  "172.18.0.2:8089"
]
Node2 peers:
[
  "172.18.0.4:8088",
  "172.18.0.3:8090"
]
Node3 peers:
[
  "172.18.0.2:8089",
  "172.18.0.4:8088"
]

2. Removing Node3 from Node1 and Node2's peer lists...
Node3 ID: 172.18.0.3:8090

3. Verifying Node3 was removed from peer lists:
Node1 peers:
[
  "172.18.0.2:8089"
]
Node2 peers:
[
  "172.18.0.4:8088"
]

4. Testing counter propagation with removed peer...
Incrementing counter on Node1...
{"count":2,"node_id":"172.18.0.4:8088"}

5. Checking if counter updated on Node2 but not on Node3:
Node1 count:
{
  "count": 2,
  "node_id": "172.18.0.4:8088"
}
Node2 count:
{
  "count": 2,
  "node_id": "172.18.0.2:8089"
}
Node3 count:
{
  "count": 1,
  "node_id": "172.18.0.3:8090"
}

6. Re-adding Node3 to restore network...

7. Verifying Node3 was added back:
Node1 peers:
[
  "172.18.0.3:8090",
  "172.18.0.2:8089"
]
Node2 peers:
[
  "172.18.0.4:8088",
  "172.18.0.3:8090"
]

8. Explicitly syncing Node3 with the current counter value...
Current count from Node1: 2

9. Verifying all nodes have the same counter value:
Node1 count:
{
  "count": 2,
  "node_id": "172.18.0.4:8088"
}
Node2 count:
{
  "count": 2,
  "node_id": "172.18.0.2:8089"
}
Node3 count:
{
  "count": 2,
  "node_id": "172.18.0.3:8090"
}

✅ Existing peer removal test completed!
(base) shubhamgautam@shubhams-MacBook-Air assignment-advance % make test-count               
🔍 Fetching count from all nodes...
Node1:
{
  "count": 2,
  "node_id": "172.18.0.4:8088"
}
Node2:
{
  "count": 2,
  "node_id": "172.18.0.2:8089"
}
Node3:
{
  "count": 2,
  "node_id": "172.18.0.3:8090"
}
✅ Count API test completed!
(base) shubhamgautam@shubhams-MacBook-Air assignment-advance % make test-all-peers           
🔍 Listing peers for all nodes...
Node1 peers:
[
  "172.18.0.3:8090",
  "172.18.0.2:8089"
]
Node2 peers:
[
  "172.18.0.4:8088",
  "172.18.0.3:8090"
]
Node3 peers:
[
  "172.18.0.4:8088",
  "172.18.0.2:8089"
]
✅ All nodes peers listed!
