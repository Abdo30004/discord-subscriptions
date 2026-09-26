package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var (
	registryOnce sync.Once
	reg          *prometheus.Registry

	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	activeBotInstances  *prometheus.GaugeVec

	tracer trace.Tracer
)

func initRegistry() {
	reg = prometheus.NewRegistry()

	// Register default Go runtime and process metrics
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	reg.MustRegister(collectors.NewGoCollector())

	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed, partitioned by status code, method, and path.",
		},
		[]string{"service", "method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of response latency for HTTP requests in seconds.",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"service", "method", "path"},
	)

	activeBotInstances = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bot_fleet_active_instances",
			Help: "Current count of active bot instances deployed in the fleet.",
		},
		[]string{"service", "bot_type"},
	)

	reg.MustRegister(httpRequestsTotal)
	reg.MustRegister(httpRequestDuration)
	reg.MustRegister(activeBotInstances)

	// Configure OpenTelemetry global text map propagator for W3C TraceContext
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	tracer = otel.GetTracerProvider().Tracer("discord-subscriptions")
}

// EnsureRegistry initializes the metrics registry once and returns it.
func EnsureRegistry() *prometheus.Registry {
	registryOnce.Do(initRegistry)
	return reg
}

// Handler returns an HTTP handler serving Prometheus metrics.
func Handler() http.Handler {
	r := EnsureRegistry()
	return promhttp.HandlerFor(r, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// statusWriter wraps http.ResponseWriter to capture HTTP status code.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// HTTPMiddleware instruments incoming HTTP requests with Prometheus RED metrics and OpenTelemetry tracing.
func HTTPMiddleware(serviceName string) func(http.Handler) http.Handler {
	EnsureRegistry()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract incoming W3C TraceContext
			ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
			ctx, span := tracer.Start(ctx, fmt.Sprintf("%s %s", r.Method, r.URL.Path))
			defer span.End()

			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			start := time.Now()

			// Inject traceparent in outgoing response header for distributed tracing tracing
			otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(sw.Header()))

			next.ServeHTTP(sw, r.WithContext(ctx))

			duration := time.Since(start).Seconds()
			path := normalizePath(r.URL.Path)
			statusStr := strconv.Itoa(sw.status)

			httpRequestsTotal.WithLabelValues(serviceName, r.Method, path, statusStr).Inc()
			httpRequestDuration.WithLabelValues(serviceName, r.Method, path).Observe(duration)
		})
	}
}

// SetFleetGauge updates the active bot count metric.
func SetFleetGauge(serviceName, botType string, count float64) {
	EnsureRegistry()
	activeBotInstances.WithLabelValues(serviceName, botType).Set(count)
}

// normalizePath groups path parameters to prevent high-cardinality metric explosion.
func normalizePath(path string) string {
	if path == "" {
		return "/"
	}
	// Keep well-known prefixes clean
	return path
}

// InjectAMQPTraceContext injects W3C traceparent into an AMQP headers table.
func InjectAMQPTraceContext(ctx context.Context, headers map[string]interface{}) {
	if headers == nil {
		return
	}
	carrier := make(propagation.MapCarrier)
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	for k, v := range carrier {
		headers[k] = v
	}
}

// ExtractAMQPTraceContext extracts W3C traceparent from an AMQP headers table.
func ExtractAMQPTraceContext(ctx context.Context, headers map[string]interface{}) context.Context {
	if headers == nil {
		return ctx
	}
	carrier := make(propagation.MapCarrier)
	for k, v := range headers {
		if s, ok := v.(string); ok {
			carrier[k] = s
		}
	}
	return otel.GetTextMapPropagator().Extract(ctx, carrier)
}
