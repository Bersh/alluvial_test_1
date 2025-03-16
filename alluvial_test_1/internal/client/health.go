package client

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bersh/alluvial_test_1/internal/metrics"
)

// CheckAllHealth checks the health of all clients
func (p *PoolStruct) CheckAllHealth() {
	log.Println("Running health check for all clients...")

	clients := p.GetAllClients()
	for _, client := range clients {
		go func(c *Client) {
			isHealthy := p.CheckClientHealth(c)
			p.SetClientAvailability(c.Name, isHealthy)
			log.Printf("Client %s health: %v\n", c.Name, isHealthy)
		}(client)
	}
}

// CheckClientHealth checks if a specific client is healthy
func (p *PoolStruct) CheckClientHealth(client *Client) bool {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_blockNumber",
		"params":  []interface{}{},
		"id":      1,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling health check payload for %s: %v\n", client.Name, err)
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", client.URL, strings.NewReader(string(payloadBytes)))
	if err != nil {
		log.Printf("Error creating health check request for %s: %v\n", client.Name, err)
		return false
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		log.Printf("Health check request failed for %s: %v\n", client.Name, err)
		metrics.RecordClientError(client.Name, "health_check")
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Health check received non-200 status code from %s: %d\n", client.Name, resp.StatusCode)
		return false
	}

	// Read and parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Error parsing health check response from %s: %v\n", client.Name, err)
		return false
	}

	// Check if there's an error in the response
	if _, ok := response["error"]; ok {
		log.Printf("Health check received error response from %s: %v\n", client.Name, response["error"])
		return false
	}

	return true
}
