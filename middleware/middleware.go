package middleware

import (
	"net/http"
)

// Middleware interface
type Middleware interface {
	Handler(h http.Handler) http.Handler
}
