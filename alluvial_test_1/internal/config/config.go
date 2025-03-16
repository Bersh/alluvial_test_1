package config

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// Config holds the application configuration
type Config struct {
	ServerPort          string
	RequestTimeout      time.Duration
	HealthCheckInterval time.Duration
	Clients             []ClientConfig
}

// ClientConfig holds configuration for a single Ethereum client
type ClientConfig struct {
	URL     string
	Name    string
	Timeout time.Duration
}

// Load loads the application configuration from environment variables
func Load() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	requestTimeout := 15 * time.Second
	healthCheckInterval := 30 * time.Second

	clients := getClientConfigsFromEnv()
	if len(clients) == 0 {
		return nil, errors.New("no Ethereum clients configured. Please set ETH_CLIENT_<N>_URL and ETH_CLIENT_<N>_NAME in .env")
	}

	return &Config{
		ServerPort:          port,
		RequestTimeout:      requestTimeout,
		HealthCheckInterval: healthCheckInterval,
		Clients:             clients,
	}, nil
}

// getClientConfigsFromEnv retrieves Ethereum client configurations from environment variables
func getClientConfigsFromEnv() []ClientConfig {
	var clients []ClientConfig

	for i := 1; ; i++ {
		urlKey := fmt.Sprintf("ETH_CLIENT_%d_URL", i)
		nameKey := fmt.Sprintf("ETH_CLIENT_%d_NAME", i)

		url := os.Getenv(urlKey)
		name := os.Getenv(nameKey)

		if url == "" {
			break
		}

		if name == "" {
			name = fmt.Sprintf("client-%d", i)
		}

		clients = append(clients, ClientConfig{
			URL:     url,
			Name:    name,
			Timeout: 10 * time.Second,
		})
	}

	return clients
}
