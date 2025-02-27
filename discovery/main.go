package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// Node represents a registered service node
type Node struct {
	ID          string    `json:"id"`
	Address     string    `json:"address"`
	ServicePort string    `json:"service_port"`
	LastSeen    time.Time `json:"last_seen"`
}

var (
	registry    = make(map[string]*Node)
	registryMu  sync.RWMutex
	servicePort string
)

func main() {
	// Get port from environment variables
	servicePort = os.Getenv("PORT")
	if servicePort == "" {
		log.Println("❌ PORT not set, using default 8000")
		servicePort = "8000"
	}

	log.Printf("🚀 Starting Discovery Service on port %s\n", servicePort)

	// Start regular health check for registered nodes
	go performHealthChecks()

	// Setup HTTP handlers
	http.HandleFunc("/register", registerNodeHandler)
	http.HandleFunc("/unregister", unregisterNodeHandler)
	http.HandleFunc("/nodes", getNodesHandler)
	http.HandleFunc("/health", healthCheckHandler)

	// Start HTTP server
	serverAddr := fmt.Sprintf(":%s", servicePort)
	log.Fatal(http.ListenAndServe(serverAddr, nil))
}

// registerNodeHandler handles node registration requests
func registerNodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var node Node
	if err := json.NewDecoder(r.Body).Decode(&node); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if node.ID == "" || node.Address == "" || node.ServicePort == "" {
		http.Error(w, "Missing required fields: id, address, service_port", http.StatusBadRequest)
		return
	}

	// Update last seen timestamp
	node.LastSeen = time.Now()

	// Add to registry
	registryMu.Lock()
	registry[node.ID] = &node
	registryMu.Unlock()

	log.Printf("✅ Registered node: %s at %s:%s\n", node.ID, node.Address, node.ServicePort)

	// Return success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Node registered successfully",
	})
}

// unregisterNodeHandler handles node unregistration requests
func unregisterNodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if request.ID == "" {
		http.Error(w, "Missing required field: id", http.StatusBadRequest)
		return
	}

	// Remove from registry
	registryMu.Lock()
	delete(registry, request.ID)
	registryMu.Unlock()

	log.Printf("🔌 Unregistered node: %s\n", request.ID)

	// Return success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Node unregistered successfully",
	})
}

// getNodesHandler returns all registered nodes
func getNodesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters
	query := r.URL.Query()
	excludeID := query.Get("exclude")

	// Build response with all nodes except the excluded one
	registryMu.RLock()
	nodes := make([]*Node, 0, len(registry))
	for id, node := range registry {
		if id != excludeID {
			nodes = append(nodes, node)
		}
	}
	registryMu.RUnlock()

	// Return nodes list
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodes)
}

// healthCheckHandler provides a simple health check endpoint
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "discovery",
	})
}

// performHealthChecks periodically checks if registered nodes are still alive
func performHealthChecks() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("🩺 Performing health checks on registered nodes")

		// Get current time to check for timeout
		now := time.Now()
		timeout := 30 * time.Second

		// List of nodes to remove (dead nodes)
		var nodesToRemove []string

		// Check all nodes
		registryMu.RLock()
		for id, node := range registry {
			// Check if node hasn't been seen recently
			if now.Sub(node.LastSeen) > timeout {
				nodesToRemove = append(nodesToRemove, id)
				continue
			}

			// Try to ping node's health endpoint
			url := fmt.Sprintf("http://%s:%s/health", node.Address, node.ServicePort)
			client := http.Client{Timeout: 2 * time.Second}
			_, err := client.Get(url)
			if err != nil {
				log.Printf("⚠️ Health check failed for node %s: %v\n", id, err)
				nodesToRemove = append(nodesToRemove, id)
			}
		}
		registryMu.RUnlock()

		// Remove dead nodes
		if len(nodesToRemove) > 0 {
			registryMu.Lock()
			for _, id := range nodesToRemove {
				log.Printf("❌ Removing dead node: %s\n", id)
				delete(registry, id)
			}
			registryMu.Unlock()
		}
	}
}
