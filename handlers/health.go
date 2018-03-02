package handlers

import (
	"context"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type HealthMonitor struct {
	Pinger  Pinger
	Timeout time.Duration
	Logger  log.FieldLogger
}

func (h *HealthMonitor) log() log.FieldLogger {
	if h.Logger != nil {
		return h.Logger
	}
	return log.StandardLogger()
}

func (h *HealthMonitor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, _ := context.WithTimeout(context.Background(), h.Timeout)
	err := h.Pinger.Ping(ctx)

	response := struct {
		IsAlive bool   `json:"is_alive"`
		Error   string `json:"error,omitempty"`
	}{
		IsAlive: err == nil,
	}

	if err != nil {
		h.log().WithField("err", err).Println("Auth backend health monitor")
		response.Error = err.Error()
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var code int
	if err != nil {
		code = http.StatusServiceUnavailable
	} else {
		code = http.StatusOK
	}

	w.WriteHeader(code)

	json.NewEncoder(w).Encode(&response)
}
