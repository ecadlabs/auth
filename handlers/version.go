package handlers

import (
	"encoding/json"
	"net/http"
)

type VersionHandler string

func (h VersionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := struct {
		Version string `json:"version"`
	}{
		Version: string(h),
	}

	json.NewEncoder(w).Encode(response)
	return
}
