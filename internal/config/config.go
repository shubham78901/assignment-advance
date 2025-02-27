package config

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// Config holds the application configuration
type Config struct {
	ServicePort      string
	DiscoveryPort    int
	NodeID           string
	PeerList         string
	DiscoveryEnabled bool
}

// Initialize sets up the configuration from environment variables and command-line arguments
func Initialize() *Config {
	cfg := &Config{
		DiscoveryPort:    8089,
		DiscoveryEnabled: true,
	}

	// Get port from environment variables
	cfg.ServicePort = os.Getenv("PORT")
	if cfg.ServicePort == "" {
		fmt.Println("❌ PORT not set, using default 8088")
		cfg.ServicePort = "8088"
	}

	// Generate a unique node ID based on IP and port
	cfg.NodeID = getNodeID(cfg.ServicePort)
	fmt.Printf("🆔 Node ID: %s\n", cfg.NodeID)

	// Get peer list from environment variables
	cfg.PeerList = os.Getenv("PEERS")

	// Parse command-line arguments
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--discovery-port=") {
			portStr := strings.TrimPrefix(arg, "--discovery-port=")
			fmt.Sscanf(portStr, "%d", &cfg.DiscoveryPort)
		} else if arg == "--disable-discovery" {
			cfg.DiscoveryEnabled = false
		}
	}

	return cfg
}

// getNodeID generates a unique ID for this node
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
