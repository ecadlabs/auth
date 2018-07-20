package utils

import (
	"encoding/json"
	"net/http"
	"strings"

	"git.ecadlabs.com/ecad/auth/errors"
)

func JSONError(w http.ResponseWriter, err string, code errors.Code) {
	response := errors.Response{
		Code: code,
	}

	status := response.HTTPStatus()

	if err != "" {
		response.Error = err
	} else {
		response.Error = http.StatusText(status)
	}

	JSONResponse(w, status, &response)
}

func JSONErrorResponse(w http.ResponseWriter, err error) {
	res := errors.ErrorResponse(err)
	JSONResponse(w, res.HTTPStatus(), res)
}

func JSONResponse(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
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

// Lazy email syntax verification
func ValidEmail(s string) bool {
	i := strings.IndexByte(s, '@')
	return i >= 1 && i < len(s)-1 && i == strings.LastIndexByte(s, '@')
}
