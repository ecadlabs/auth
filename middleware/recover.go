package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	log "github.com/sirupsen/logrus"
)

type Recover struct {
	Logger log.FieldLogger
}

func (r *Recover) log() log.FieldLogger {
	if r.Logger != nil {
		return r.Logger
	}
	return log.StandardLogger()
}

func (r *Recover) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			err := recover()
			if err == nil {
				return
			}
			stack := debug.Stack()

			fields := log.Fields{
				"method": req.Method,
				"path":   req.URL.Path,
			}
			r.log().WithFields(fields).Println(string(stack))

			http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		}()

		h.ServeHTTP(w, req)
	})
}
