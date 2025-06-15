package middleware

import (
	"net/http"
)

type Middleware struct {
}

var middlewareInstance *Middleware

func NewMiddleware() *Middleware {
	if middlewareInstance != nil {
		return middlewareInstance
	}

	newMiddleware := &Middleware{}
	middlewareInstance = newMiddleware

	return middlewareInstance
}

func (m *Middleware) Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// defer drainBuffer(r.Body)

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
