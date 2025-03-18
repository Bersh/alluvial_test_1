package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	RequestDuration    *prometheus.HistogramVec
	RequestTotal       *prometheus.CounterVec
	ClientErrors       *prometheus.CounterVec
	ClientAvailability *prometheus.GaugeVec
	BalanceDiscrepancy *prometheus.CounterVec
}

// Global metrics instance - can be nil in test environments
var M *Metrics

// Init initializes and registers all Prometheus metrics
func Init() {
	M = &Metrics{
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "request_duration_seconds",
				Help:    "Time (in seconds) spent serving HTTP requests",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint", "status"},
		),
		RequestTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Count of all HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		ClientErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "client_errors_total",
				Help: "Count of errors per client",
			},
			[]string{"client_name", "error_type"},
		),
		ClientAvailability: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "client_availability",
				Help: "Availability status of each client (1=available, 0=unavailable)",
			},
			[]string{"client_name"},
		),
		BalanceDiscrepancy: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "balance_discrepancy_total",
				Help: "Count of balance discrepancies between clients",
			},
			[]string{"address"},
		),
	}

	prometheus.MustRegister(
		M.RequestDuration,
		M.RequestTotal,
		M.ClientErrors,
		M.ClientAvailability,
		M.BalanceDiscrepancy,
	)
}

func RecordClientError(clientName, errorType string) {
	if M == nil || M.ClientErrors == nil {
		return
	}
	M.ClientErrors.WithLabelValues(clientName, errorType).Inc()
}

func SetClientAvailability(clientName string, isAvailable bool) {
	if M == nil || M.ClientAvailability == nil {
		return
	}

	var value float64 = 0
	if isAvailable {
		value = 1
	}
	M.ClientAvailability.WithLabelValues(clientName).Set(value)
}

func RecordBalanceDiscrepancy(address string) {
	if M == nil || M.BalanceDiscrepancy == nil {
		return
	}
	M.BalanceDiscrepancy.WithLabelValues(address).Inc()
}
