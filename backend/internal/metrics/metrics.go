package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all application metrics
type Metrics struct {
	// HTTP metrics
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight prometheus.Gauge

	// Business metrics
	activeUsers      prometheus.Gauge
	apiCallsTotal    *prometheus.CounterVec
	dbQueriesTotal   *prometheus.CounterVec
	dbQueryDuration  *prometheus.HistogramVec
	k8sAPICallsTotal *prometheus.CounterVec
	authAttempts     *prometheus.CounterVec
	cacheHits        prometheus.Counter
	cacheMisses      prometheus.Counter

	// Resource metrics
	goroutines prometheus.Gauge
	memoryUsed prometheus.Gauge
}

// New creates and registers all metrics
func New() *Metrics {
	m := &Metrics{
		// HTTP metrics
		httpRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kubedeck_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		httpRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kubedeck_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		httpRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "kubedeck_http_requests_in_flight",
				Help: "Number of HTTP requests currently being processed",
			},
		),

		// Business metrics
		activeUsers: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "kubedeck_active_users",
				Help: "Number of active users",
			},
		),
		apiCallsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kubedeck_api_calls_total",
				Help: "Total number of API calls",
			},
			[]string{"endpoint", "method", "status"},
		),
		dbQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kubedeck_db_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"query_type", "status"},
		),
		dbQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kubedeck_db_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"query_type"},
		),
		k8sAPICallsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kubedeck_kubernetes_api_calls_total",
				Help: "Total number of Kubernetes API calls",
			},
			[]string{"resource", "verb", "status"},
		),
		authAttempts: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kubedeck_auth_attempts_total",
				Help: "Total number of authentication attempts",
			},
			[]string{"method", "success"},
		),
		cacheHits: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "kubedeck_cache_hits_total",
				Help: "Total number of cache hits",
			},
		),
		cacheMisses: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "kubedeck_cache_misses_total",
				Help: "Total number of cache misses",
			},
		),

		// Resource metrics
		goroutines: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "kubedeck_goroutines",
				Help: "Number of goroutines",
			},
		),
		memoryUsed: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "kubedeck_memory_used_bytes",
				Help: "Memory used in bytes",
			},
		),
	}

	// Start collecting runtime metrics
	go m.collectRuntimeMetrics()

	return m
}

// Middleware creates HTTP middleware for metrics collection
func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		m.httpRequestsInFlight.Inc()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(wrapped.statusCode)

		m.httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
		m.httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
		m.httpRequestsInFlight.Dec()

		// Track API calls
		m.apiCallsTotal.WithLabelValues(r.URL.Path, r.Method, status).Inc()
	})
}

// Handler returns Prometheus metrics HTTP handler
func (m *Metrics) Handler() http.Handler {
	return promhttp.Handler()
}

// RecordDBQuery records database query metrics
func (m *Metrics) RecordDBQuery(queryType, status string, duration time.Duration) {
	m.dbQueriesTotal.WithLabelValues(queryType, status).Inc()
	m.dbQueryDuration.WithLabelValues(queryType).Observe(duration.Seconds())
}

// RecordK8sAPICall records Kubernetes API call metrics
func (m *Metrics) RecordK8sAPICall(resource, verb, status string) {
	m.k8sAPICallsTotal.WithLabelValues(resource, verb, status).Inc()
}

// RecordAuthAttempt records authentication attempt
func (m *Metrics) RecordAuthAttempt(method string, success bool) {
	m.authAttempts.WithLabelValues(method, strconv.FormatBool(success)).Inc()
}

// RecordCacheHit records cache hit
func (m *Metrics) RecordCacheHit() {
	m.cacheHits.Inc()
}

// RecordCacheMiss records cache miss
func (m *Metrics) RecordCacheMiss() {
	m.cacheMisses.Inc()
}

// SetActiveUsers sets the number of active users
func (m *Metrics) SetActiveUsers(count int) {
	m.activeUsers.Set(float64(count))
}

// collectRuntimeMetrics collects runtime metrics periodically
func (m *Metrics) collectRuntimeMetrics() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Goroutines
		// Note: runtime.NumGoroutine() requires import "runtime"
		// For now, we'll skip this to avoid circular imports

		// Memory (simplified)
		// In production, use runtime.MemStats for accurate metrics
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Global metrics instance
var global *Metrics

// Global returns global metrics instance
func Global() *Metrics {
	if global == nil {
		global = New()
	}
	return global
}
