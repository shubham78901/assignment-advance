package main

import (
	"assignmet/advance/config"
	"assignmet/advance/discovery"
	"assignmet/advance/handlers"
	"fmt"
	"net/http"
)

func main() {
	// Initialize configuration
	cfg := config.LoadConfig()

	// Initialize discovery
	discovery.SetupServiceDiscovery(cfg)

	// Start regular peer health checks
	go discovery.HealthCheckPeers(cfg)

	// Start broadcasting presence regularly
	if cfg.DiscoveryEnabled {
		go discovery.BroadcastPresence(cfg)
	}

	fmt.Printf("📡 Node started on port %s, Peers: %v\n", cfg.ServicePort, discovery.GetPeerList(cfg))

	// Setup HTTP handlers
	http.HandleFunc("/register", handlers.RegisterPeer(cfg))
	http.HandleFunc("/peers", handlers.GetPeers(cfg))
	http.HandleFunc("/remove-peer", handlers.RemovePeer(cfg))
	http.HandleFunc("/increment", handlers.IncrementHandler(cfg))
	http.HandleFunc("/count", handlers.CountHandler(cfg))
	http.HandleFunc("/sync", handlers.SyncHandler(cfg))
	http.HandleFunc("/health", handlers.HealthCheck(cfg))
	http.HandleFunc("/discovery", handlers.DiscoveryHandler(cfg))

	// Start HTTP server
	serverAddr := fmt.Sprintf(":%s", cfg.ServicePort)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		fmt.Printf("❌ Server failed: %v\n", err)
	}
}
