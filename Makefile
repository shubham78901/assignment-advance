.PHONY: build start stop restart logs test-health test-sync test-increment test-count test-peers test-all-peers test-full remove-containers

# Updated host ports to match docker-compose mappings:
# node1 is accessible at host port 9088, node2 at 9089, node3 at 9090.
PORT_1=9088
PORT_2=9089
PORT_3=9090

make :stop build  start

build:
	@echo "🔨 Building Docker images..."
	docker-compose build

start:
	@echo "🚀 Starting all nodes..."
	docker-compose up -d



stop:
	@echo "🛑 Stopping all nodes..."
	docker-compose down

restart: stop start

logs:
	@echo "📜 Showing logs..."
	docker-compose logs -f

test-health:
	@echo "🩺 Checking health of all nodes..."
	@echo "Node1:"; curl -s http://localhost:$(PORT_1)/health || echo "Health check failed for node1"
	@echo "Node2:"; curl -s http://localhost:$(PORT_2)/health || echo "Health check failed for node2"
	@echo "Node3:"; curl -s http://localhost:$(PORT_3)/health || echo "Health check failed for node3"
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
	@echo "Node1:"; curl -s http://localhost:$(PORT_1)/count | jq || echo "Count check failed for node1"
	@echo "Node2:"; curl -s http://localhost:$(PORT_2)/count | jq || echo "Count check failed for node2"
	@echo "Node3:"; curl -s http://localhost:$(PORT_3)/count | jq || echo "Count check failed for node3"
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

test-full: test-health test-increment test-sync test-count test-peers test-all-peers
	@echo "🎯 Full test sequence completed!"

remove-containers:
	@echo "🗑️ Removing containers for node1, node2, and node3..."
	@docker rm -f node1 node2 node3 || echo "No such containers found"
