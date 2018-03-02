package middleware

import (
	"net/http"
)

// ResponseWriter wraps http.ResponseWriter to save HTTP status code
type ResponseWriter struct {
	http.ResponseWriter
	Status int
}

func (rw *ResponseWriter) WriteHeader(s int) {
	rw.Status = s
	rw.ResponseWriter.WriteHeader(s)
}

func (rw *ResponseWriter) Write(data []byte) (int, error) {
	if rw.Status == 0 {
		rw.Status = http.StatusOK
	}

	return rw.ResponseWriter.Write(data)
}
