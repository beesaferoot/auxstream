package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auxstream_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auxstream_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	SearchRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auxstream_search_requests_total",
			Help: "Total number of search requests",
		},
		[]string{"source", "status"},
	)

	SearchDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auxstream_search_duration_seconds",
			Help:    "Search request duration in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"source"},
	)

	CacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auxstream_cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_type"},
	)

	CacheMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auxstream_cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_type"},
	)

	StreamTokensGenerated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auxstream_stream_tokens_generated_total",
			Help: "Total number of stream tokens generated",
		},
	)

	StreamTokensValidated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auxstream_stream_tokens_validated_total",
			Help: "Total number of stream token validations",
		},
		[]string{"result"},
	)

	IndexerJobsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auxstream_indexer_jobs_total",
			Help: "Total number of indexer jobs run",
		},
	)

	IndexerTracksIndexed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auxstream_indexer_tracks_indexed_total",
			Help: "Total number of tracks indexed",
		},
		[]string{"source", "status"},
	)

	IndexerJobDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "auxstream_indexer_job_duration_seconds",
			Help:    "Indexer job duration in seconds",
			Buckets: []float64{1, 5, 10, 30, 60, 120, 300, 600},
		},
	)

	DatabaseQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auxstream_database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "status"},
	)

	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auxstream_database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"operation"},
	)

	RateLimitExceeded = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auxstream_rate_limit_exceeded_total",
			Help: "Total number of rate limit exceeded events",
		},
		[]string{"limit_type"},
	)

	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "auxstream_active_connections",
			Help: "Number of active connections",
		},
	)

	TracksUploaded = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auxstream_tracks_uploaded_total",
			Help: "Total number of tracks uploaded",
		},
	)

	AuthAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auxstream_auth_attempts_total",
			Help: "Total number of authentication attempts",
		},
		[]string{"method", "status"},
	)
)

func RecordHTTPRequest(method, endpoint, status string, duration float64) {
	HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}

func RecordSearchRequest(source, status string, duration float64) {
	SearchRequestsTotal.WithLabelValues(source, status).Inc()
	SearchDuration.WithLabelValues(source).Observe(duration)
}

func RecordCacheHit(cacheType string) {
	CacheHits.WithLabelValues(cacheType).Inc()
}

func RecordCacheMiss(cacheType string) {
	CacheMisses.WithLabelValues(cacheType).Inc()
}

func RecordStreamToken() {
	StreamTokensGenerated.Inc()
}

func RecordStreamTokenValidation(valid bool) {
	result := "invalid"
	if valid {
		result = "valid"
	}
	StreamTokensValidated.WithLabelValues(result).Inc()
}

func RecordIndexerJob(duration float64, successCount, failCount int) {
	IndexerJobsTotal.Inc()
	IndexerJobDuration.Observe(duration)
	if successCount > 0 {
		IndexerTracksIndexed.WithLabelValues("all", "success").Add(float64(successCount))
	}
	if failCount > 0 {
		IndexerTracksIndexed.WithLabelValues("all", "failed").Add(float64(failCount))
	}
}

func RecordIndexerJobSource(source string, successCount, failCount float64) {
	if successCount > 0 {
		IndexerTracksIndexed.WithLabelValues(source, "success").Add(successCount)
	}
	if failCount > 0 {
		IndexerTracksIndexed.WithLabelValues(source, "failed").Add(failCount)
	}
}

func RecordDatabaseQuery(operation, status string, duration float64) {
	DatabaseQueriesTotal.WithLabelValues(operation, status).Inc()
	DatabaseQueryDuration.WithLabelValues(operation).Observe(duration)
}

func RecordRateLimitExceeded(limitType string) {
	RateLimitExceeded.WithLabelValues(limitType).Inc()
}

func IncActiveConnections() {
	ActiveConnections.Inc()
}

func DecActiveConnections() {
	ActiveConnections.Dec()
}

func RecordTrackUpload() {
	TracksUploaded.Inc()
}

func RecordAuthAttempt(method, status string) {
	AuthAttemptsTotal.WithLabelValues(method, status).Inc()
}
