package metrics

import (
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type recorderState struct {
	enabled bool
	service string
}

var (
	state recorderState
	once  sync.Once

	httpRequestsTotal          *prometheus.CounterVec
	httpRequestDurationSeconds *prometheus.HistogramVec
	cacheHitsTotal             *prometheus.CounterVec
	cacheMissesTotal           *prometheus.CounterVec
	writeBehindLagSeconds      *prometheus.GaugeVec
	writeBehindQueueLength     *prometheus.GaugeVec
	writeBehindBatchSize       *prometheus.GaugeVec
	writeBehindBatchDuration   *prometheus.HistogramVec
	writeBehindErrorsTotal     *prometheus.CounterVec
)

// Init configures the shared Prometheus collectors.
func Init(service string, enabled bool) {
	once.Do(func() {
		state = recorderState{
			enabled: enabled,
			service: service,
		}
		if !enabled {
			return
		}
		httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests served by the API",
		}, []string{"service", "method", "path", "status"})

		httpRequestDurationSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Latency of HTTP handlers in seconds",
			Buckets: prometheus.DefBuckets,
		}, []string{"service", "method", "path"})

		cacheHitsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		}, []string{"service", "status"})

		cacheMissesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses and bypasses",
		}, []string{"service", "status"})

		writeBehindLagSeconds = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "write_behind_lag_seconds",
			Help: "Age of the oldest message processed in the latest batch",
		}, []string{"service"})

		writeBehindQueueLength = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "write_behind_queue_length",
			Help: "Length of the write-behind Redis stream",
		}, []string{"service"})

		writeBehindBatchSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "write_behind_batch_size",
			Help: "Number of events processed in the latest batch",
		}, []string{"service"})

		writeBehindBatchDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "write_behind_batch_duration_seconds",
			Help:    "Processing latency of write-behind batches",
			Buckets: prometheus.DefBuckets,
		}, []string{"service"})

		writeBehindErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "write_behind_errors_total",
			Help: "Total number of write-behind errors",
		}, []string{"service"})
	})
}

func enabled() bool {
	return state.enabled
}

// ObserveHTTPRequest records request counts and latency.
func ObserveHTTPRequest(method, path string, status int, duration time.Duration) {
	if !enabled() {
		return
	}
	labels := []string{state.service, method, path, strconv.Itoa(status)}
	httpRequestsTotal.WithLabelValues(labels...).Inc()
	httpRequestDurationSeconds.WithLabelValues(state.service, method, path).Observe(duration.Seconds())
}

// RecordCacheStatus tracks cache behaviour for a single lookup.
func RecordCacheStatus(status string) {
	if !enabled() {
		return
	}
	switch status {
	case "fresh", "stale":
		cacheHitsTotal.WithLabelValues(state.service, status).Inc()
	default:
		cacheMissesTotal.WithLabelValues(state.service, status).Inc()
	}
}

// RecordWriteBehindBatch captures queue stats per processed batch.
func RecordWriteBehindBatch(batchSize int, duration time.Duration, lagSeconds float64, queueLength int64) {
	if !enabled() {
		return
	}
	writeBehindBatchSize.WithLabelValues(state.service).Set(float64(batchSize))
	writeBehindBatchDuration.WithLabelValues(state.service).Observe(duration.Seconds())
	if lagSeconds >= 0 {
		writeBehindLagSeconds.WithLabelValues(state.service).Set(lagSeconds)
	}
	if queueLength >= 0 {
		writeBehindQueueLength.WithLabelValues(state.service).Set(float64(queueLength))
	}
}

// RecordWriteBehindError increments the write-behind error counter.
func RecordWriteBehindError() {
	if !enabled() {
		return
	}
	writeBehindErrorsTotal.WithLabelValues(state.service).Inc()
}
