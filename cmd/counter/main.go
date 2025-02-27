package main

import (
	"assignmet/advance/internal/config"
	"assignmet/advance/internal/counter"
	"assignmet/advance/internal/discovery"
	"assignmet/advance/internal/handlers"
	"assignmet/advance/internal/peers"
	"fmt"
	"net/http"
	"os"
)

func main() {
	// Initialize configuration
	cfg := config.Initialize()

	// Initialize counter service
	counterService := counter.New()

	// Initialize peers manager
	peersManager := peers.New(cfg.NodeID)

	// Setup initial peers from environment or command line
	peersManager.SetupInitialPeers(cfg.PeerList, os.Args)

	// Initialize service discovery if enabled
	if cfg.DiscoveryEnabled {
		discoveryService := discovery.New(cfg.NodeID, cfg.DiscoveryPort, peersManager)
		go discoveryService.Listen()
		go discoveryService.BroadcastPresence()

		// Register with any peers provided in the initial configuration
		for _, peer := range peersManager.GetPeerList() {
			go peersManager.RegisterWithPeer(peer, discoveryService)
		}

		fmt.Printf("🔍 Service discovery enabled on port %d\n", cfg.DiscoveryPort)
	} else {
		fmt.Println("ℹ️ Service discovery is disabled")
	}

	// Start regular peer health checks
	go peersManager.HealthCheckPeers()

	fmt.Printf("📡 Node started on port %s, Peers: %v\n", cfg.ServicePort, peersManager.GetPeerList())

	// Setup HTTP handlers
	handler := handlers.New(cfg.NodeID, counterService, peersManager)

	// Register all HTTP routes
	http.HandleFunc("/register", handler.RegisterPeer)
	http.HandleFunc("/peers", handler.GetPeers)
	http.HandleFunc("/remove-peer", handler.RemovePeer)
	http.HandleFunc("/increment", handler.IncrementHandler)
	http.HandleFunc("/count", handler.CountHandler)
	http.HandleFunc("/sync", handler.SyncHandler)
	http.HandleFunc("/health", handler.HealthCheck)
	http.HandleFunc("/discovery", handler.DiscoveryHandler)

	// Start HTTP server
	serverAddr := fmt.Sprintf(":%s", cfg.ServicePort)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		fmt.Printf("❌ Server failed: %v\n", err)
	}
}
