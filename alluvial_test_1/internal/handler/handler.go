package handler

import (
	"github.com/go-chi/chi/v5"
	"net/http"

	"github.com/bersh/alluvial_test_1/internal/client"
	"github.com/bersh/alluvial_test_1/internal/config"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupRouter configures the HTTP router
func SetupRouter(clientPool *client.PoolStruct, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(PrometheusMiddleware)

	balanceHandler := NewBalanceHandler(clientPool, cfg.RequestTimeout)
	healthHandler := NewHealthHandler(clientPool)

	r.Get("/eth/balance/{address}", balanceHandler.GetBalance)

	r.Get("/health/live", healthHandler.LivenessCheck)
	r.Get("/health/ready", healthHandler.ReadinessCheck)

	r.Handle("/metrics", promhttp.Handler())

	return r
}
