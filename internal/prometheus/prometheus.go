package prometheus

import (
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Total HTTP 5xx responses",
		},
		[]string{"method", "path"},
	)
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{
			ResponseWriter: w,
			status:         200,
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		method := r.Method
		path := normalizePath(r.URL.Path)
		status := strconv.Itoa(rw.status)

		httpDuration.WithLabelValues(method, path, status).Observe(duration)
		httpRequestsTotal.WithLabelValues(method, path, status).Inc()

		if rw.status >= 500 {
			httpErrorsTotal.WithLabelValues(method, path).Inc()
		}
	})
}

var (
	userIDRe      = regexp.MustCompile(`^/internal/v1/users/\d+$`)
	usersBatchRe  = regexp.MustCompile(`^/internal/v1/users/batch$`)
	profileRe     = regexp.MustCompile(`^/api/v1/profile$`)
	profileListRe = regexp.MustCompile(`^/api/v1/profile/list$`)
)

func normalizePath(path string) string {
	switch {
	case profileRe.MatchString(path):
		return "/api/v1/profile"
	case profileListRe.MatchString(path):
		return "/api/v1/profile/list"
	case userIDRe.MatchString(path):
		return "/internal/v1/users/{id}"
	case usersBatchRe.MatchString(path):
		return "/internal/v1/users/batch"
	}
	return path
}