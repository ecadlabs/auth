package middleware

import (
	"net/http"
)

const (
	//TokenContextKey Context value key for request token
	TokenContextKey = "token"
)

// ResponseWriter wraps http.ResponseWriter to save HTTP status code
type ResponseStatusWriter interface {
	http.ResponseWriter
	Status() int
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

type responseWriterHijacker struct {
	*responseWriter
	http.Hijacker
}

func NewResponseStatusWriter(w http.ResponseWriter) ResponseStatusWriter {
	ret := &responseWriter{
		ResponseWriter: w,
	}

	if h, ok := w.(http.Hijacker); ok {
		return &responseWriterHijacker{
			responseWriter: ret,
			Hijacker:       h,
		}
	}

	return ret
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(s int) {
	rw.status = s
	rw.ResponseWriter.WriteHeader(s)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}

	return rw.ResponseWriter.Write(data)
}

var _ http.ResponseWriter = &responseWriter{}
var _ http.ResponseWriter = &responseWriterHijacker{}
var _ http.Hijacker = &responseWriterHijacker{}
