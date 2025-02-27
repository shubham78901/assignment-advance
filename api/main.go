package main

import (
	"assignmet/advance/api/config"
	"assignmet/advance/api/gossip"
	"assignmet/advance/api/handlers"
	"assignmet/advance/api/peers"
	"fmt"
	"net/http"
	"os"
)

func main() {
	port, peerList := config.LoadConfig()
	fmt.Println("🚀 Server starting on port:", port)

	// Add each peer to the peers package
	for _, peer := range peerList {
		peers.AddPeer(peer)
	}

	fmt.Println("🌐 Initial peers:", peers.GetPeers())

	// Start peer health check in the background
	go peers.HealthCheckPeers()
	go gossip.RegisterWithPeers(port)

	// Define HTTP routes
	http.HandleFunc("/register", handlers.RegisterPeer)
	http.HandleFunc("/peers", handlers.GetPeers)
	http.HandleFunc("/remove-peer", handlers.RemovePeer)
	http.HandleFunc("/increment", handlers.IncrementHandler)
	http.HandleFunc("/count", handlers.CountHandler)
	http.HandleFunc("/sync", handlers.SyncHandler)
	http.HandleFunc("/health", handlers.HealthCheck)

	// Start the server
	serverAddr := fmt.Sprintf(":%s", port)
	fmt.Printf("📡 Server running on port %s\n", port)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		fmt.Printf("❌ Server failed: %v\n", err)
		os.Exit(1)
	}
}
