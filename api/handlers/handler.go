package handlers

import (
	"assignmet/advance/api/peers"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// RegisterPeer handles peer registration safely.
func RegisterPeer(w http.ResponseWriter, r *http.Request) {
	var peer struct {
		ID string `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&peer); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if peer.ID != "" {
		peers.AddPeer(peer.ID) // ✅ Thread-safe peer addition
		fmt.Printf("✅ Registered peer: %s\n", peer.ID)
		go func() {
			peersList := peers.GetPeers() // Take a **snapshot** of peers to avoid modifying it concurrently
			var wg sync.WaitGroup
			for _, p := range peersList {
				wg.Add(1)
				go func(p string) {
					defer wg.Done()
					fmt.Printf("🔄 Notifying %s about new peer %s\n", p, peer.ID)
					// Safe peer notification logic here...
				}(p)
			}
			wg.Wait()
		}()
	} else {
		fmt.Println("⚠️ Received empty peer ID")
	}

	json.NewEncoder(w).Encode(peers.GetPeers())
}
