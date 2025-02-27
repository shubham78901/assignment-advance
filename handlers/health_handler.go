package handlers

import (
	"assignmet/advance/config"
	"encoding/json"

	"net/http"
)

func HealthCheck(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"node_id": cfg.NodeID,
		})
	}
}
