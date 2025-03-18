package handler

import (
	"encoding/json"
	"net/http"

	"github.com/bersh/alluvial_test_1/internal/client"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	clientPool *client.PoolStruct
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(clientPool *client.PoolStruct) *HealthHandler {
	return &HealthHandler{
		clientPool: clientPool,
	}
}

// LivenessCheck handles the liveness probe
func (h *HealthHandler) LivenessCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
}

// ReadinessCheck handles the readiness probe
func (h *HealthHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	available := h.clientPool.HasAvailableClients()

	if !available {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "not ready",
			"message": "no clients available",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}
