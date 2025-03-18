package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bersh/alluvial_test_1/internal/client"
	"github.com/bersh/alluvial_test_1/internal/config"
	"github.com/bersh/alluvial_test_1/internal/handler"
	"github.com/bersh/alluvial_test_1/internal/metrics"
	"github.com/bersh/alluvial_test_1/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file:", err)
	}

	log.Println("Starting Ethereum Balance Proxy service...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	metrics.Init()

	clientPool, err := client.NewPool(cfg.Clients)
	if err != nil {
		log.Fatalf("Failed to initialize client pool: %v", err)
	}

	go clientPool.CheckAllHealth()

	go func() {
		ticker := time.NewTicker(cfg.HealthCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				clientPool.CheckAllHealth()
			}
		}
	}()

	router := handler.SetupRouter(clientPool, cfg)

	srv := server.New(router, cfg.ServerPort)
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server started on port %s\n", cfg.ServerPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server gracefully stopped")
}
