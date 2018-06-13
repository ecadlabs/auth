package handlers

import (
	"context"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type HealthMonitor struct {
	Pinger  Pinger
	Timeout time.Duration
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
		log.Println(err)
		response.Error = err.Error()
	}

	var code int
	if err != nil {
		code = http.StatusServiceUnavailable
	} else {
		code = http.StatusOK
	}

	JSONResponse(w, code, &response)
}
