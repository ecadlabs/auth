package middleware

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"strconv"
	"time"
)

type Prometheus struct {
	counter *prometheus.CounterVec
	hist    *prometheus.HistogramVec
}

func NewPrometheus(n ...string) *Prometheus {
	var (
		namePrefix string
		helpPrefix string
	)

	if len(n) != 0 {
		helpPrefix = n[0] + ": "
		namePrefix = n[0]

		if namePrefix[len(namePrefix)-1] != '_' {
			namePrefix += "_"
		}
	}

	labels := []string{"code", "method", "path"}

	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: namePrefix + "http_requests_total",
		Help: helpPrefix + "Total number of HTTP requests",
	}, labels)

	if err := prometheus.Register(counter); err != nil {
		// Reuse collector
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			counter = are.ExistingCollector.(*prometheus.CounterVec)
		} else {
			panic(err)
		}
	}

	hist := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: namePrefix + "http_request_duration_milliseconds",
		Help: helpPrefix + "HTTP request duration",
		// TODO buckets
	}, labels)

	if err := prometheus.Register(hist); err != nil {
		// Reuse collector
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			hist = are.ExistingCollector.(*prometheus.HistogramVec)
		} else {
			panic(err)
		}
	}

	return &Prometheus{
		counter: counter,
		hist:    hist,
	}
}

func (p *Prometheus) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := ResponseWriter{ResponseWriter: w}

		timestamp := time.Now()
		next.ServeHTTP(&rw, r)
		duration := time.Since(timestamp)

		labels := prometheus.Labels{
			"code":   strconv.FormatInt(int64(rw.Status), 10),
			"method": r.Method,
			"path":   r.URL.Path,
		}

		p.counter.With(labels).Inc()
		p.hist.With(labels).Observe(float64(duration / time.Millisecond))
	})
}
