package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"dapodik-readonly-api/internal/response"
)

func APIKey(expectedKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			providedKey := r.Header.Get("X-API-Key")
			if providedKey == "" {
				if authorization := r.Header.Get("Authorization"); strings.HasPrefix(authorization, "Bearer ") {
					providedKey = strings.TrimPrefix(authorization, "Bearer ")
				}
			}

			if len(providedKey) != len(expectedKey) || subtle.ConstantTimeCompare([]byte(providedKey), []byte(expectedKey)) != 1 {
				response.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
