package handlers

import (
	"net/http"
)

type VersionHandler string

func (h VersionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := struct {
		Version string `json:"version"`
	}{
		Version: string(h),
	}

	JSONResponse(w, http.StatusOK, &response)
}
