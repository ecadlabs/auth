package middleware

// Logging middleware inspired by github.com/urfave/negroni

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// Logging is a logrus-enabled logging middleware
type Logging struct {
	Logger log.FieldLogger
}

func (l *Logging) log() log.FieldLogger {
	if l.Logger != nil {
		return l.Logger
	}
	return log.StandardLogger()
}

// Handler wraps provided http.Handler with middleware
func (l *Logging) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timestamp := time.Now()

		rw := NewResponseStatusWriter(w)
		h.ServeHTTP(rw, r)

		fields := log.Fields{
			"start_time": timestamp.Format(time.RFC3339),
			"duration":   time.Since(timestamp),
			"status":     rw.Status(),
			"hostname":   r.Host,
			"method":     r.Method,
			"path":       r.URL.Path,
		}

		l.log().WithFields(fields).Println(r.Method + " " + r.URL.Path)
	})
}
