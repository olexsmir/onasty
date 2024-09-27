package metrics

import (
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
)

func RecordSuccessfulRequestMetric(method, uri string) {
	successfulHTTPRequest.With(prometheus.Labels{"method": method, "uri": uri}).Inc()
}

func RecordFailedRequestMetric(method, uri string) {
	failedHTTPRequest.With(prometheus.Labels{"method": method, "uri": uri}).Inc()
}
