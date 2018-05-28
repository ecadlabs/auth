package handlers

import (
	"encoding/json"
	"net/http"
)

func JSONError(w http.ResponseWriter, err string, code int) {
	response := struct {
		Error string `json:"error,omitempty"`
	}{
		Error: err,
	}

	JSONResponse(w, code, &response)
}

func JSONResponse(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

type Paginated struct {
	Value interface{} `json:"value"`
	Next  string      `json:"next,omitempty"`
}
