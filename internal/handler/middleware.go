package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bersh/alluvial_test_1/internal/metrics"
	"github.com/go-chi/chi/v5/middleware"
)

// PrometheusMiddleware collects HTTP metrics
func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		duration := time.Since(start).Seconds()

		endpoint := r.URL.Path

		status := fmt.Sprintf("%d", ww.Status())
		metrics.M.RequestDuration.WithLabelValues(r.Method, endpoint, status).Observe(duration)
		metrics.M.RequestTotal.WithLabelValues(r.Method, endpoint, status).Inc()
	})
}
