package handlers

import (
	"net/http"

	"github.com/ecadlabs/auth/utils"
)

type VersionHandler string

func (h VersionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := struct {
		Version string `json:"version"`
	}{
		Version: string(h),
	}

	utils.JSONResponse(w, http.StatusOK, &response)
}
