package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/ecadlabs/auth/utils"
	log "github.com/sirupsen/logrus"
)

//Pinger interface representing a service that can ping something
type Pinger interface {
	Ping(ctx context.Context) error
}

//HealthMonitor service that do health check
type HealthMonitor struct {
	Pinger  Pinger
	Timeout time.Duration
}

func (h *HealthMonitor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), h.Timeout)
	defer cancel()

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

	utils.JSONResponse(w, code, &response)
}
