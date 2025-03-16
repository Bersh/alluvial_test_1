package handler

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"

	"github.com/bersh/alluvial_test_1/internal/client"
	"github.com/bersh/alluvial_test_1/internal/service"
	"github.com/ethereum/go-ethereum/common"
)

// BalanceHandler handles balance-related endpoints
type BalanceHandler struct {
	clientPool     *client.PoolStruct
	requestTimeout time.Duration
	balanceService *service.BalanceService
}

// NewBalanceHandler creates a new balance handler
func NewBalanceHandler(clientPool *client.PoolStruct, requestTimeout time.Duration) *BalanceHandler {
	balanceService := service.NewBalanceService(clientPool)
	return &BalanceHandler{
		clientPool:     clientPool,
		requestTimeout: requestTimeout,
		balanceService: balanceService,
	}
}

// GetBalance handles the eth_getBalance proxy endpoint
func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	address := chi.URLParam(r, "address")

	// Validate the Ethereum address
	if !common.IsHexAddress(address) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid Ethereum address"})
		return
	}

	// Normalize the address
	address = common.HexToAddress(address).Hex()

	blockParam := r.URL.Query().Get("block")
	if blockParam == "" {
		blockParam = "latest"
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.requestTimeout)
	defer cancel()

	balance, err := h.balanceService.GetBalance(ctx, address, blockParam)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"balance": balance.String()})
}
