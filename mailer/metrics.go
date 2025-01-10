package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	emailSentSuccessfully = promauto.NewCounter(prometheus.CounterOpts{ //nolint:exhaustruct
		Name: "mail_sent_total",
		Help: "the total number of successfully sent email",
	})

	emailFailedToSend = promauto.NewCounterVec(prometheus.CounterOpts{ //nolint:exhaustruct
		Name: "mail_failed_total",
		Help: "the total number of email that failed to send",
	}, []string{"request_id"})
)

func MetricsHandler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.Handler())
	return mux
}

func RecordEmailSent() {
	go emailSentSuccessfully.Inc()
}

func RecordEmailFailed(reqid string) {
	go emailFailedToSend.With(prometheus.Labels{
		"request_id": reqid,
	}).Inc()
}
