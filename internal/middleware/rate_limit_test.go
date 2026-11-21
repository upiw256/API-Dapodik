package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimitIsSharedAcrossRoutesForEachIP(t *testing.T) {
	now := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
	limiter := &rateLimiter{clients: make(map[string]requestWindow), maxRequests: 5, window: 10 * time.Minute, idleReset: time.Minute, now: func() time.Time { return now }}
	var handled int
	handler := limiter.middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { handled++; w.WriteHeader(http.StatusOK) }))

	for i := 0; i < 5; i++ {
		path := "/api/health"
		if i%2 == 1 {
			path = "/api/rombel"
		}
		r := httptest.NewRequest(http.MethodGet, path, nil)
		r.RemoteAddr = "192.0.2.10:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("request %d status=%d, want 200", i+1, w.Code)
		}
	}

	r := httptest.NewRequest(http.MethodGet, "/api/peserta-didik", nil)
	r.RemoteAddr = "192.0.2.10:23456"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("sixth request status=%d, want 429", w.Code)
	}
	if got := w.Header().Get("Retry-After"); got != "60" {
		t.Fatalf("Retry-After=%q, want 60", got)
	}
	if handled != 5 {
		t.Fatalf("handler called %d times, want 5", handled)
	}

	otherIP := httptest.NewRequest(http.MethodGet, "/api/rombel", nil)
	otherIP.RemoteAddr = "192.0.2.11:12345"
	otherResponse := httptest.NewRecorder()
	handler.ServeHTTP(otherResponse, otherIP)
	if otherResponse.Code != http.StatusOK {
		t.Fatalf("different IP status=%d, want 200", otherResponse.Code)
	}
}

func TestRateLimitWindowExpiresAndOptionsIsIgnored(t *testing.T) {
	now := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
	limiter := &rateLimiter{clients: make(map[string]requestWindow), maxRequests: 5, window: 10 * time.Minute, idleReset: time.Minute, now: func() time.Time { return now }}
	handler := limiter.middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	for i := 0; i < 5; i++ {
		r := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		r.RemoteAddr = "192.0.2.20:12345"
		handler.ServeHTTP(httptest.NewRecorder(), r)
	}

	options := httptest.NewRequest(http.MethodOptions, "/api/health", nil)
	options.RemoteAddr = "192.0.2.20:12345"
	optionsResponse := httptest.NewRecorder()
	handler.ServeHTTP(optionsResponse, options)
	if optionsResponse.Code != http.StatusNoContent {
		t.Fatalf("OPTIONS status=%d, want 204", optionsResponse.Code)
	}

	now = now.Add(time.Minute)
	r := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	r.RemoteAddr = "192.0.2.20:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusNoContent {
		t.Fatalf("request after window status=%d, want 204", w.Code)
	}
	if got := w.Header().Get("Retry-After"); got != "" {
		t.Fatalf("unexpected Retry-After: %s", got)
	}
}
