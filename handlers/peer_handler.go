package handlers

import (
	"assignmet/advance/config"
	"assignmet/advance/discovery"
	"encoding/json"
	"fmt"

	"net/http"
)

func RegisterPeer(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var peer struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&peer); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Don't add ourselves as a peer
		if peer.ID == cfg.NodeID {
			w.WriteHeader(http.StatusOK)
			return
		}

		discovery.PeersMu.Lock()
		wasNew := !cfg.Peers[peer.ID]
		cfg.Peers[peer.ID] = true
		discovery.PeersMu.Unlock()

		if wasNew {
			fmt.Printf("✅ Registered new peer: %s\n", peer.ID)
		}

		w.WriteHeader(http.StatusOK)
	}
}
