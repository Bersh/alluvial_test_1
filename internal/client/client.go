package client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bersh/alluvial_test_1/internal/config"
	"github.com/bersh/alluvial_test_1/internal/metrics"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"io"
	"log"
	"math/big"
	"net/http"
	"strings"
	"sync"
)

// PoolStruct defines the interface for the client pool
//
//go:generate mockery --name Pool
type Pool interface {
	QueryBalanceFromAllClients(ctx context.Context, address, blockParam string) ([]BalanceResponse, error)
	GetAvailableClients() []*Client
	HasAvailableClients() bool
	SetClientAvailability(clientName string, isAvailable bool)
	GetAllClients() []*Client
	CheckClientHealth(client *Client) bool
	CheckAllHealth()
}

// Client represents an Ethereum JSON-RPC client
type Client struct {
	URL         string
	Name        string
	HTTPClient  *http.Client
	IsAvailable bool
}

// PoolStruct manages multiple Ethereum clients
type PoolStruct struct {
	clients      []*Client
	clientsMutex sync.RWMutex
}

// NewClient creates a new Ethereum client
func NewClient(cfg config.ClientConfig) *Client {
	return &Client{
		URL:         cfg.URL,
		Name:        cfg.Name,
		HTTPClient:  &http.Client{Timeout: cfg.Timeout},
		IsAvailable: true,
	}
}

// NewPool creates a pool of Ethereum clients
func NewPool(clientConfigs []config.ClientConfig) (*PoolStruct, error) {
	if len(clientConfigs) == 0 {
		return nil, fmt.Errorf("no client configurations provided")
	}

	clients := make([]*Client, 0, len(clientConfigs))
	for _, cfg := range clientConfigs {
		clients = append(clients, NewClient(cfg))
	}

	return &PoolStruct{
		clients: clients,
	}, nil
}

// HasAvailableClients checks if there's at least one available client
func (p *PoolStruct) HasAvailableClients() bool {
	p.clientsMutex.RLock()
	defer p.clientsMutex.RUnlock()

	for _, client := range p.clients {
		if client.IsAvailable {
			return true
		}
	}
	return false
}

// GetAvailableClients returns all available clients
func (p *PoolStruct) GetAvailableClients() []*Client {
	p.clientsMutex.RLock()
	defer p.clientsMutex.RUnlock()

	available := make([]*Client, 0)
	for _, client := range p.clients {
		if client.IsAvailable {
			available = append(available, client)
		}
	}
	return available
}

// SetClientAvailability sets the availability status of a client
func (p *PoolStruct) SetClientAvailability(clientName string, isAvailable bool) {
	p.clientsMutex.Lock()
	defer p.clientsMutex.Unlock()

	for _, client := range p.clients {
		if client.Name == clientName {
			client.IsAvailable = isAvailable
			metrics.SetClientAvailability(client.Name, isAvailable)
			break
		}
	}
}

// QueryBalance queries the balance from a specific client
func (c *Client) QueryBalance(ctx context.Context, address, blockParam string) (*big.Int, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getBalance",
		"params":  []interface{}{address, blockParam},
		"id":      1,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.URL, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		metrics.RecordClientError(c.Name, "request_failed")
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Error closing body: %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		metrics.RecordClientError(c.Name, "non_200_status")
		return nil, fmt.Errorf("received non-200 status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		JSONRPC string `json:"jsonrpc"`
		ID      int    `json:"id"`
		Result  string `json:"result"`
		Error   struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		metrics.RecordClientError(c.Name, "parse_error")
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	if result.Error.Message != "" {
		metrics.RecordClientError(c.Name, "rpc_error")
		return nil, fmt.Errorf("RPC error: %s (code: %d)", result.Error.Message, result.Error.Code)
	}

	balance, err := hexutil.DecodeBig(result.Result)
	if err != nil {
		metrics.RecordClientError(c.Name, "decode_error")
		return nil, fmt.Errorf("error parsing balance: %w", err)
	}

	return balance, nil
}

// GetAllClients returns all clients in the pool
func (p *PoolStruct) GetAllClients() []*Client {
	p.clientsMutex.RLock()
	defer p.clientsMutex.RUnlock()

	clients := make([]*Client, len(p.clients))
	copy(clients, p.clients)

	return clients
}

// BalanceResponse represents a balance response from a client
type BalanceResponse struct {
	ClientName string
	Balance    *big.Int
	Error      error
}

// QueryBalanceFromAllClients queries all available clients for balance
func (p *PoolStruct) QueryBalanceFromAllClients(ctx context.Context, address, blockParam string) ([]BalanceResponse, error) {
	clients := p.GetAvailableClients()
	if len(clients) == 0 {
		return nil, fmt.Errorf("no Ethereum clients available")
	}

	resultChan := make(chan BalanceResponse, len(clients))

	for _, client := range clients {
		go func(c *Client) {
			balance, err := c.QueryBalance(ctx, address, blockParam)
			resultChan <- BalanceResponse{
				ClientName: c.Name,
				Balance:    balance,
				Error:      err,
			}
		}(client)
	}

	responses := make([]BalanceResponse, 0, len(clients))
	for i := 0; i < len(clients); i++ {
		resp := <-resultChan
		if resp.Error == nil {
			responses = append(responses, resp)
		} else {
			log.Printf("Error from client %s: %v\n", resp.ClientName, resp.Error)
		}
	}

	if len(responses) == 0 {
		return nil, fmt.Errorf("failed to retrieve balance from any client")
	}

	return responses, nil
}
