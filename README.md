

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

# Run Build containers and start

[make]

# Run To see all logs

[make logs]


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




............................................CONSOLE-LOGS..................................................................


(base) shubhamgautam@shubhams-MacBook-Air assignment-advance % make test-full
🩺 Checking health of all nodes...
Node1:
{
  "node_id": "172.18.0.4:8088",
  "status": "ok"
}
Node2:
{
  "node_id": "172.18.0.2:8089",
  "status": "ok"
}
Node3:
{
  "node_id": "172.18.0.3:8090",
  "status": "ok"
}
✅ Health check completed!
📈 Incrementing counter on Node1...
{"count":1,"node_id":"172.18.0.4:8088"}
✅ Increment API test completed!
🔄 Testing sync API on all nodes...
Syncing node1 with count 1
Syncing node2 with count 1
Syncing node3 with count 1
✅ Sync API test completed!
🔍 Fetching count from all nodes...
Node1:
{
  "count": 1,
  "node_id": "172.18.0.4:8088"
}
Node2:
{
  "count": 1,
  "node_id": "172.18.0.2:8089"
}
Node3:
{
  "count": 1,
  "node_id": "172.18.0.3:8090"
}
✅ Count API test completed!
🔍 Listing peers for all nodes...
Node1 peers:
[
  "172.18.0.3:8090",
  "172.18.0.2:8089"
]
Node2 peers:
[
  "172.18.0.3:8090",
  "172.18.0.4:8088"
]
Node3 peers:
[
  "172.18.0.4:8088",
  "172.18.0.2:8089"
]
✅ All nodes peers listed!
🧹 Cleaning up any previous test nodes...
🔍 Testing service discovery...

1. Checking discovery endpoint for Node1:
{
  "node_id": "172.18.0.4:8088",
  "peers": [
    "172.18.0.3:8090",
    "172.18.0.2:8089"
  ]
}

2. Testing auto-discovery by adding a new node dynamically...
Starting a new container discovery-test-node without explicitly connecting it to others...
7eaf1c9c3102f95d6d784e305646ed20c062d94715b9af7e6f4e5390a21c4b8d

3. Waiting for discovery to propagate (15 seconds)...

4. Checking if the new node was discovered by existing nodes:
Node1 peers:
[
  "172.18.0.3:8090",
  "172.18.0.2:8089",
  "172.18.0.5:8091"
]
Node2 peers:
[
  "172.18.0.3:8090",
  "172.18.0.4:8088",
  "172.18.0.5:8091"
]
Node3 peers:
[
  "172.18.0.4:8088",
  "172.18.0.2:8089",
  "172.18.0.5:8091"
]

5. Checking if the new node discovered existing nodes:
Test node peers:
[
  "172.18.0.4:8088",
  "172.18.0.3:8090",
  "172.18.0.2:8089"
]

6. Testing counter propagation to the new node...
Incrementing counter on Node1...
{"count":2,"node_id":"172.18.0.4:8088"}

Checking counter value on all nodes including the new one:
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
Test node count:
{
  "count": 2,
  "node_id": "172.18.0.5:8091"
}

7. Cleaning up the dynamically added node...
discovery-test-node
discovery-test-node

✅ Service discovery test completed!
🔍 Testing removal of an existing peer...
1. Checking current peers on all nodes:
Node1 peers:
[
  "172.18.0.5:8091",
  "172.18.0.3:8090",
  "172.18.0.2:8089"
]
Node2 peers:
[
  "172.18.0.3:8090",
  "172.18.0.4:8088",
  "172.18.0.5:8091"
]
Node3 peers:
[
  "172.18.0.2:8089",
  "172.18.0.5:8091",
  "172.18.0.4:8088"
]

2. Removing Node3 from Node1 and Node2's peer lists...
Node3 ID: 172.18.0.3:8090

3. Verifying Node3 was removed from peer lists:
Node1 peers:
[
  "172.18.0.2:8089",
  "172.18.0.5:8091"
]
Node2 peers:
[
  "172.18.0.5:8091",
  "172.18.0.4:8088"
]

4. Testing counter propagation with removed peer...
Incrementing counter on Node1...
{"count":3,"node_id":"172.18.0.4:8088"}

5. Checking if counter updated on Node2 but not on Node3:
Node1 count:
{
  "count": 3,
  "node_id": "172.18.0.4:8088"
}
Node2 count:
{
  "count": 3,
  "node_id": "172.18.0.2:8089"
}
Node3 count:
{
  "count": 2,
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
  "172.18.0.3:8090",
  "172.18.0.4:8088"
]

8. Explicitly syncing Node3 with the current counter value...
Current count from Node1: 3

9. Verifying all nodes have the same counter value:
Node1 count:
{
  "count": 3,
  "node_id": "172.18.0.4:8088"
}
Node2 count:
{
  "count": 3,
  "node_id": "172.18.0.2:8089"
}
Node3 count:
{
  "count": 3,
  "node_id": "172.18.0.3:8090"
}

✅ Existing peer removal test completed!
🎯 Full test sequence completed!
(base) shubhamgautam@shubhams-MacBook-Air assignment-advance % 











<img width="1470" alt="Screenshot 2025-02-26 at 12 33 30 PM" src="https://github.com/user-attachments/assets/b5d14ce9-e502-4d5b-8df2-45b5ba1cd80f" />
<img width="1470" alt="Screenshot 2025-02-27 at 8 27 43 PM" src="https://github.com/user-attachments/assets/210aae72-43b7-4396-aa5a-12d6fa60945f" />

<img width="1470" alt="Screenshot 2025-02-27 at 8 28 09 PM" src="https://github.com/user-attachments/assets/a209d0b5-c7bd-428c-93da-3086a770ac35" />
<img width="1470" alt="Screenshot 2025-02-27 at 8 28 04 PM" src="https://github.com/user-attachments/assets/1716de47-6067-4e0b-a27e-fe3ee40c0eb0" />
