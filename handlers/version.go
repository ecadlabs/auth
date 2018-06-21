package handlers

import (
	"git.ecadlabs.com/ecad/auth/utils"
	"net/http"
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
