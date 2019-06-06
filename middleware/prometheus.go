package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

type Prometheus struct {
	counter      *prometheus.CounterVec
	hist         *prometheus.HistogramVec
	useHandlerID bool
}

func NewPrometheus() *Prometheus {
	return NewPrometheusFull("", false)
}

func NewPrometheusWithHandlerID() *Prometheus {
	return NewPrometheusFull("", true)
}

func NewPrometheusFull(prefix string, handlerID bool) *Prometheus {
	var helpPrefix string

	if prefix != "" {
		helpPrefix = prefix + ": "

		if prefix[len(prefix)-1] != '_' {
			prefix += "_"
		}
	}

	labels := []string{"code", "method" /*, "path"*/}
	if handlerID {
		labels = append(labels, "handler")
	}

	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: prefix + "http_requests_total",
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
		Name:    prefix + "http_request_duration_milliseconds",
		Help:    helpPrefix + "HTTP request duration",
		Buckets: prometheus.ExponentialBuckets(250, 2, 6),
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
		counter:      counter,
		hist:         hist,
		useHandlerID: handlerID,
	}
}

func (p *Prometheus) handler(w http.ResponseWriter, r *http.Request, handler string, next http.Handler) {
	rw := NewResponseStatusWriter(w)

	timestamp := time.Now()
	next.ServeHTTP(rw, r)
	duration := time.Since(timestamp)

	labels := prometheus.Labels{
		"code":   strconv.FormatInt(int64(rw.Status()), 10),
		"method": r.Method,
	}

	if p.useHandlerID {
		labels["handler"] = handler
	}

	p.counter.With(labels).Inc()
	p.hist.With(labels).Observe(float64(duration / time.Millisecond))
}

func (p *Prometheus) HandlerWithID(handler string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.handler(w, r, handler, next)
	})
}

func (p *Prometheus) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var handler string
		if route := mux.CurrentRoute(r); route != nil {
			if tpl, err := route.GetPathTemplate(); err == nil {
				handler = tpl
			}
		}
		p.handler(w, r, handler, next)
	})
}
