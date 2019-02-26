package utils

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/ecadlabs/auth/errors"
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

func GetRemoteAddr(r *http.Request) string {
	if fh := r.Header.Get("Forwarded"); fh != "" {
		chunks := strings.Split(fh, ",")

		for _, c := range chunks {
			opts := strings.Split(strings.TrimSpace(c), ";")

			for _, o := range opts {
				v := strings.SplitN(strings.TrimSpace(o), "=", 2)

				if len(v) == 2 && v[0] == "for" {
					if addr := strings.Trim(v[1], "\"[]"); addr != "" {
						return addr
					}
				}
			}
		}
	}

	if xfh := r.Header.Get("X-Forwarded-For"); xfh != "" {
		chunks := strings.Split(xfh, ",")
		for _, c := range chunks {
			if c = strings.Trim(strings.TrimSpace(c), "\"[]"); c != "" {
				return c
			}
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return r.RemoteAddr
}
