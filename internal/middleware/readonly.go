package middleware

import (
	"net/http"

	"dapodik-readonly-api/internal/response"
)

func ReadOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
		default:
			w.Header().Set("Allow", "GET, HEAD, OPTIONS")
			response.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})
}
