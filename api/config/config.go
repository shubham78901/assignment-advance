package config

import (
	"fmt"
	"os"
	"strings"
)

// LoadConfig initializes configuration settings and returns peers.
func LoadConfig() (string, []string) {
	port := os.Getenv("PORT")
	if port == "" {
		fmt.Println("❌ PORT not set, using default 8088")
		port = "8088"
	}

	peersEnv := os.Getenv("PEERS") // Comma-separated list of peers
	var peersList []string
	if peersEnv != "" {
		peersList = strings.Split(peersEnv, ",")
	}

	return port, peersList
}
