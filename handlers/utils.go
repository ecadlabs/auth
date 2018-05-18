package handlers

import (
	"encoding/json"
	"net/http"
)

func JSONError(w http.ResponseWriter, err string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)

	response := struct {
		Error string `json:"error,omitempty"`
	}{
		Error: err,
	}
	json.NewEncoder(w).Encode(&response)
}
