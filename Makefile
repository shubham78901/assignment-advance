.PHONY: build start stop restart logs test-health test-sync test-increment test-count test-peers test-all-peers test-full remove-containers test-discovery clean

# Updated host ports to match docker-compose mappings:
# node1 is accessible at host port 9088, node2 at 9089, node3 at 9090.
PORT_1=9088
PORT_2=9089
PORT_3=9090
TEST_NODE_PORT=9091
TEST_NODE_NAME=discovery-test-node

make: stop build start

build:
	@echo "🔨 Building Docker images..."
	docker-compose build

start:
	@echo "🚀 Starting all nodes..."
	docker-compose up -d
	@echo "⏳ Waiting for service discovery to initialize..."
	@sleep 10

stop:
	@echo "🛑 Stopping all nodes..."
	docker-compose down

restart: stop start

logs:
	@echo "📜 Showing logs..."
	docker-compose logs -f

test-health:
	@echo "🩺 Checking health of all nodes..."
	@echo "Node1:"; curl -s http://localhost:$(PORT_1)/health | jq . || echo "Health check failed for node1"
	@echo "Node2:"; curl -s http://localhost:$(PORT_2)/health | jq . || echo "Health check failed for node2"
	@echo "Node3:"; curl -s http://localhost:$(PORT_3)/health | jq . || echo "Health check failed for node3"
	@echo "✅ Health check completed!"

test-sync:
	@echo "🔄 Testing sync API on all nodes..."
	@{ \
	  count1=`curl -s http://localhost:$(PORT_1)/count | jq '.count'`; \
	  echo "Syncing node1 with count $$count1"; \
	  curl -X POST -H "Content-Type: application/json" -d "{\"count\": $$count1}" http://localhost:$(PORT_1)/sync; \
	  count2=`curl -s http://localhost:$(PORT_2)/count | jq '.count'`; \
	  echo "Syncing node2 with count $$count2"; \
	  curl -X POST -H "Content-Type: application/json" -d "{\"count\": $$count2}" http://localhost:$(PORT_2)/sync; \
	  count3=`curl -s http://localhost:$(PORT_3)/count | jq '.count'`; \
	  echo "Syncing node3 with count $$count3"; \
	  curl -X POST -H "Content-Type: application/json" -d "{\"count\": $$count3}" http://localhost:$(PORT_3)/sync; \
	}; \
	sleep 2; \
	echo "✅ Sync API test completed!"

test-increment:
	@echo "📈 Incrementing counter on Node1..."
	@curl -X POST http://localhost:$(PORT_1)/increment
	@sleep 2
	@echo "✅ Increment API test completed!"

test-count:
	@echo "🔍 Fetching count from all nodes..."
	@echo "Node1:"; curl -s http://localhost:$(PORT_1)/count | jq . || echo "Count check failed for node1"
	@echo "Node2:"; curl -s http://localhost:$(PORT_2)/count | jq . || echo "Count check failed for node2"
	@echo "Node3:"; curl -s http://localhost:$(PORT_3)/count | jq . || echo "Count check failed for node3"
	@echo "✅ Count API test completed!"

test-peers:
	@echo "🔍 Testing /peers endpoint on Node1 after manual registration..."
	@echo "Registering a peer 'test-peer' on Node1..."
	@curl -X POST -H "Content-Type: application/json" -d '{"id": "test-peer"}' http://localhost:$(PORT_1)/register
	@sleep 1
	@echo "\nListing peers for Node1:"
	@curl -s http://localhost:$(PORT_1)/peers | jq .
	@echo "\n✅ /peers test completed!"

test-all-peers:
	@echo "🔍 Listing peers for all nodes..."
	@echo "Node1 peers:"; curl -s http://localhost:$(PORT_1)/peers | jq . || echo "Failed to get peers for node1"
	@echo "Node2 peers:"; curl -s http://localhost:$(PORT_2)/peers | jq . || echo "Failed to get peers for node2"
	@echo "Node3 peers:"; curl -s http://localhost:$(PORT_3)/peers | jq . || echo "Failed to get peers for node3"
	@echo "✅ All nodes peers listed!"

clean-test-node:
	@echo "🧹 Cleaning up any previous test nodes..."
	@docker rm -f $(TEST_NODE_NAME) 2>/dev/null || true

test-discovery: clean-test-node
	@echo "🔍 Testing service discovery..."
	@echo "\n1. Checking discovery endpoint for Node1:"
	@curl -s http://localhost:$(PORT_1)/discovery | jq .
	
	@echo "\n2. Testing auto-discovery by adding a new node dynamically..."
	@echo "Starting a new container $(TEST_NODE_NAME) without explicitly connecting it to others..."
	@NETWORK_NAME=$$(docker network ls --filter name=mynetwork --format "{{.Name}}") && \
	docker run -d --name $(TEST_NODE_NAME) --network $$NETWORK_NAME -e PORT=8091 -p $(TEST_NODE_PORT):8091 -p 9191:8089 assignment-advance-node1
	
	@echo "\n3. Waiting for discovery to propagate (15 seconds)..."
	@sleep 15
	
	@echo "\n4. Checking if the new node was discovered by existing nodes:"
	@echo "Node1 peers:"; curl -s http://localhost:$(PORT_1)/peers | jq .
	@echo "Node2 peers:"; curl -s http://localhost:$(PORT_2)/peers | jq .
	@echo "Node3 peers:"; curl -s http://localhost:$(PORT_3)/peers | jq .
	
	@echo "\n5. Checking if the new node discovered existing nodes:"
	@echo "Test node peers:"; curl -s http://localhost:$(TEST_NODE_PORT)/peers | jq .
	
	@echo "\n6. Testing counter propagation to the new node..."
	@echo "Incrementing counter on Node1..."
	@curl -X POST http://localhost:$(PORT_1)/increment
	@sleep 5
	@echo "\nChecking counter value on all nodes including the new one:"
	@echo "Node1 count:"; curl -s http://localhost:$(PORT_1)/count | jq .
	@echo "Node2 count:"; curl -s http://localhost:$(PORT_2)/count | jq .
	@echo "Node3 count:"; curl -s http://localhost:$(PORT_3)/count | jq .
	@echo "Test node count:"; curl -s http://localhost:$(TEST_NODE_PORT)/count | jq .
	
	@echo "\n7. Cleaning up the dynamically added node..."
	@docker stop $(TEST_NODE_NAME)
	@docker rm $(TEST_NODE_NAME)
	
	@echo "\n✅ Service discovery test completed!"

test-full: test-health test-increment test-sync test-count  test-all-peers test-discovery
	@echo "🎯 Full test sequence completed!"

remove-containers:
	@echo "🗑️ Removing containers..."
	@docker rm -f node1 node2 node3 2>/dev/null || true
	@docker rm -f $(TEST_NODE_NAME) 2>/dev/null || true
	@echo "✅ Containers removed!"

clean: remove-containers
	@echo "🧹 Cleaning up..."
	@docker network rm mynetwork 2>/dev/null || true
	@echo "✅ Cleanup completed!"