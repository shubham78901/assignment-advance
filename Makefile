.PHONY: build start stop restart logs test-health test-sync test-increment test-count test-peers test-all-peers test-full remove-containers

# Updated host ports to match docker-compose mappings:
PORT_1=9088
PORT_2=9089
PORT_3=9090
PORT_TEST=9091  # Added test-node port

make: stop build start

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
	@for port in $(PORT_1) $(PORT_2) $(PORT_3) $(PORT_TEST); do \
		echo "Checking node on port $$port..."; \
		curl -s http://localhost:$$port/health || echo "❌ Health check failed for port $$port"; \
	done
	@echo "✅ Health check completed for all nodes!"

test-sync:
	@echo "🔄 Testing sync API on all nodes..."
	@for port in $(PORT_1) $(PORT_2) $(PORT_3) $(PORT_TEST); do \
		count=`curl -s http://localhost:$$port/count | jq '.count'`; \
		echo "Syncing node on port $$port with count $$count"; \
		curl -X POST -H "Content-Type: application/json" -d "{\"count\": $$count}" http://localhost:$$port/sync; \
	done
	@sleep 2
	@echo "✅ Sync API test completed for all nodes!"

test-increment:
	@echo "📈 Incrementing counter on all nodes..."
	@for port in $(PORT_1) $(PORT_2) $(PORT_3) $(PORT_TEST); do \
		echo "Incrementing on port $$port..."; \
		curl -X POST http://localhost:$$port/increment; \
	done
	@sleep 2
	@echo "✅ Increment API test completed for all nodes!"

test-count:
	@echo "🔍 Fetching count from all nodes..."
	@for port in $(PORT_1) $(PORT_2) $(PORT_3) $(PORT_TEST); do \
		echo "Node on port $$port:"; \
		curl -s http://localhost:$$port/count | jq || echo "❌ Count check failed for port $$port"; \
	done
	@echo "✅ Count API test completed for all nodes!"

test-peers:
	@echo "🔍 Testing /peers endpoint..."
	@echo "➡ Registering test-node as a peer on node1..."
	@curl -s -X POST -H "Content-Type: application/json" -d '{"id": "test-node"}' http://localhost:9088/register || echo "❌ Failed to register test-node on node1"
	@sleep 1
	@echo "✅ Test-node registered on node1!"
	@echo "📜 Listing peers for all nodes..."
	@for port in $(PORT_1) $(PORT_2) $(PORT_3) $(PORT_TEST); do \
		echo "🔍 Checking peers for node at port $$port:"; \
		curl -s http://localhost:$$port/peers | jq . || echo "❌ Failed to get peers for port $$port"; \
	done
	@echo "✅ /peers test completed!"


test-all-peers:
	@echo "🔍 Listing peers for all nodes..."
	@echo "Node1 (Port: $(PORT_1)) peers:"; curl -sf http://localhost:$(PORT_1)/peers | jq . || echo "❌ Failed to get peers for Node1"
	@echo "Node2 (Port: $(PORT_2)) peers:"; curl -sf http://localhost:$(PORT_2)/peers | jq . || echo "❌ Failed to get peers for Node2"
	@echo "Node3 (Port: $(PORT_3)) peers:"; curl -sf http://localhost:$(PORT_3)/peers | jq . || echo "❌ Failed to get peers for Node3"
	@echo "Test-node (Port: $(PORT_TEST)) peers:"; curl -sf http://localhost:$(PORT_TEST)/peers | jq . || echo "❌ Failed to get peers for Test-node (Port: $(PORT_TEST))"
	@echo "✅ All nodes' peers listed!"

test-full: test-health test-increment test-sync test-count test-peers test-all-peers
	@echo "🎯 Full test sequence completed for all nodes!"

remove-containers:
	@echo "🗑️ Removing containers for all nodes..."
	@docker rm -f node1 node2 node3 test-node || echo "⚠️ No such containers found"
