// File: config/config.go
package config

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type Config struct {
	Counter          int
	ServicePort      string
	NodeID           string
	DiscoveryEnabled bool
	DiscoveryPort    int
	Peers            map[string]bool
}

func LoadConfig() *Config {
	cfg := &Config{
		Peers: make(map[string]bool),
	}

	// Get port from environment variables / command-line arguments
	cfg.ServicePort = os.Getenv("PORT")
	if cfg.ServicePort == "" {
		fmt.Println("❌ PORT not set, using default 8088")
		cfg.ServicePort = "8088"
	}

	// Generate a unique node ID based on container name and port
	// Try to use hostname first, as it's more reliable in Docker
	hostname, err := os.Hostname()
	if err == nil {
		cfg.NodeID = fmt.Sprintf("%s:%s", hostname, cfg.ServicePort)
	} else {
		// Fallback to IP-based ID
		cfg.NodeID = getNodeID(cfg.ServicePort)
	}
	fmt.Printf("🆔 Node ID: %s\n", cfg.NodeID)

	// Parse discovery port from environment
	discoveryPortStr := os.Getenv("DISCOVERY_PORT")
	if discoveryPortStr != "" {
		fmt.Sscanf(discoveryPortStr, "%d", &cfg.DiscoveryPort)
	} else {
		cfg.DiscoveryPort = 8089 // Default discovery port
	}

	// Setup initial peers from environment or command line
	setupInitialPeers(cfg)

	// Enable discovery by default
	cfg.DiscoveryEnabled = true

	return cfg
}

func getNodeID(port string) string {
	// Get the host's preferred outbound IP
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		// Fallback to local IP if cannot determine outbound IP
		return fmt.Sprintf("localhost:%s", port)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return fmt.Sprintf("%s:%s", localAddr.IP.String(), port)
}

func setupInitialPeers(cfg *Config) {
	peerList := os.Getenv("PEERS")
	if peerList != "" {
		fmt.Printf("🔄 Setting up initial peers from environment: %s\n", peerList)
		for _, peer := range strings.Split(peerList, ",") {
			if peer != cfg.NodeID {
				cfg.Peers[peer] = true
				fmt.Printf("✅ Added initial peer: %s\n", peer)
			}
		}
	}
}
