.PHONY: build start stop restart logs test-health test-discovery test-increment test-count test-peers test-all-peers test-full remove-containers clean

# Ports for service access
PORT_DISCOVERY=9000
PORT_NODE1=9088
PORT_NODE2=9089
PORT_NODE3=9090
TEST_NODE_PORT=9091
TEST_NODE_NAME=test-node

make: stop build start

build:
	@echo "🔨 Building Docker images..."
	@echo "🔨 Creating discovery.go from the discovery service code..."
	cp discovery/main.go discovery.go
	@echo "🔨 Creating node.go from the node service code..."
	cp main.go node.go
	docker-compose build


start:
	@echo "🚀 Starting all services..."
	docker-compose up -d
	@echo "⏳ Waiting for services to initialize..."
	@sleep 10

stop:
	@echo "🛑 Stopping all services..."
	docker-compose down

restart: stop start

logs:
	@echo "📜 Showing logs..."
	docker-compose logs -f

test-health:
	@echo "🩺 Checking health of all services..."
	@echo "Discovery:"; curl -s http://localhost:$(PORT_DISCOVERY)/health | jq . || echo "Health check failed for discovery"
	@echo "Node1:"; curl -s http://localhost:$(PORT_NODE1)/health | jq . || echo "Health check failed for node1"
	@echo "Node2:"; curl -s http://localhost:$(PORT_NODE2)/health | jq . || echo "Health check failed for node2"
	@echo "Node3:"; curl -s http://localhost:$(PORT_NODE3)/health | jq . || echo "Health check failed for node3"
	@echo "✅ Health check completed!"

test-discovery:
	@echo "🔍 Testing discovery service..."
	@echo "Checking registered nodes:"; curl -s http://localhost:$(PORT_DISCOVERY)/nodes | jq .
	@echo "✅ Discovery service test completed!"

test-increment:
	@echo "📈 Incrementing counter on Node1..."
	@curl -X POST http://localhost:$(PORT_NODE1)/increment
	@sleep 2
	@echo "✅ Increment API test completed!"

test-count:
	@echo "🔍 Fetching count from all nodes..."
	@echo "Node1:"; curl -s http://localhost:$(PORT_NODE1)/count | jq . || echo "Count check failed for node1"
	@echo "Node2:"; curl -s http://localhost:$(PORT_NODE2)/count | jq . || echo "Count check failed for node2"
	@echo "Node3:"; curl -s http://localhost:$(PORT_NODE3)/count | jq . || echo "Count check failed for node3"
	@echo "✅ Count API test completed!"

test-peers:
	@echo "🔍 Testing peer retrieval from nodes..."
	@echo "Node1 peers:"; curl -s http://localhost:$(PORT_NODE1)/peers | jq . || echo "Failed to get peers for node1"
	@echo "✅ Peers test completed!"

test-all-peers:
	@echo "🔍 Listing peers for all nodes..."
	@echo "Node1 peers:"; curl -s http://localhost:$(PORT_NODE1)/peers | jq . || echo "Failed to get peers for node1"
	@echo "Node2 peers:"; curl -s http://localhost:$(PORT_NODE2)/peers | jq . || echo "Failed to get peers for node2"
	@echo "Node3 peers:"; curl -s http://localhost:$(PORT_NODE3)/peers | jq . || echo "Failed to get peers for node3"
	@echo "✅ All nodes peers listed!"

create-test-node:
	@echo "🔧 Creating a test node dynamically..."
	@docker build -t test-node -f Dockerfile.node .
	@NETWORK_NAME=$$(docker network ls --filter name=mynetwork --format "{{.Name}}") && \
	docker run -d --name $(TEST_NODE_NAME) --network $$NETWORK_NAME \
		-e PORT=8091 -e DISCOVERY_URL=http://discovery:8000 -e NODE_ID=$(TEST_NODE_NAME) \
		-p $(TEST_NODE_PORT):8091 test-node
	@echo "⏳ Waiting for node to register..."
	@sleep 5
	@echo "✅ Test node created!"

test-dynamic-node: create-test-node
	@echo "🔍 Testing if discovery registered the new node..."
	@echo "Discovery nodes:"; curl -s http://localhost:$(PORT_DISCOVERY)/nodes | jq .
	@echo "Node1 peers:"; curl -s http://localhost:$(PORT_NODE1)/peers | jq .
	
	@echo "📈 Testing counter propagation to the new node..."
	@echo "Incrementing counter on Node1..."
	@curl -X POST http://localhost:$(PORT_NODE1)/increment
	@sleep 5
	@echo "\nChecking counter value on all nodes including the new one:"
	@echo "Node1 count:"; curl -s http://localhost:$(PORT_NODE1)/count | jq .
	@echo "Test node count:"; curl -s http://localhost:$(TEST_NODE_PORT)/count | jq .
	
	@echo "🧹 Removing the test node..."
	@docker stop $(TEST_NODE_NAME)
	@docker rm $(TEST_NODE_NAME)
	@echo "✅ Dynamic node test completed!"

test-full: test-health test-discovery test-increment test-count test-peers test-all-peers test-dynamic-node
	@echo "🎯 Full test sequence completed!"

remove-containers:
	@echo "🗑️ Removing containers..."
	@docker rm -f discovery node1 node2 node3 2>/dev/null || true
	@docker rm -f $(TEST_NODE_NAME) 2>/dev/null || true
	@echo "✅ Containers removed!"

clean: remove-containers
	@echo "🧹 Cleaning up..."
	@docker network rm mynetwork 2>/dev/null || true
	@rm -f discovery.go node.go 2>/dev/null || true
	@echo "✅ Cleanup completed!"