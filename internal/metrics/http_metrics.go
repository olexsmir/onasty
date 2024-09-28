package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	successfulHTTPRequest = promauto.NewCounterVec(prometheus.CounterOpts{ //nolint:exhaustruct
		Name:        "http_successful_requests_total",
		Help:        "the total number of successful http requests",
		ConstLabels: map[string]string{"status": "success"},
	}, []string{"method", "uri"})

	failedHTTPRequest = promauto.NewCounterVec(prometheus.CounterOpts{ //nolint:exhaustruct
		Name:        "http_failed_requests_total",
		Help:        "the total number of failed http requests",
		ConstLabels: map[string]string{"status": "failure"},
	}, []string{"method", "uri"})

	latencyHTTPRequest = promauto.NewHistogramVec(prometheus.HistogramOpts{ //nolint:exhaustruct
		Name:    "http_request_latency_seconds",
		Help:    "the latency of http requests in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "uri"})
)

func RecordSuccessfulRequestMetric(method, uri string) {
	go successfulHTTPRequest.With(prometheus.Labels{
		"method": method,
		"uri":    uri,
	}).Inc()
}

func RecordFailedRequestMetric(method, uri string) {
	go failedHTTPRequest.With(prometheus.Labels{
		"method": method,
		"uri":    uri,
	}).Inc()
}

func RecordLatencyRequestMetric(method, uri string, latency time.Duration) {
	go latencyHTTPRequest.With(prometheus.Labels{
		"method": method,
		"uri":    uri,
	}).Observe(latency.Seconds())
}
