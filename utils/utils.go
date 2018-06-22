package utils

import (
	"encoding/json"
	"net/http"
	"strings"
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
	Value      interface{} `json:"value"`
	TotalCount *int        `json:"total_count,omitempty"`
	Next       string      `json:"next,omitempty"`
}

func NSClaim(ns, sufix string) string {
	if strings.HasPrefix(ns, "http://") || strings.HasPrefix(ns, "https://") {
		return ns + "/" + sufix
	}

	return ns + "." + sufix
}
