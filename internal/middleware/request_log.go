package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"

	"dapodik-readonly-api/internal/response"
)

var requestSequence atomic.Uint64
var uuidInPath = regexp.MustCompile("(?i)[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}")

type statusRecorder struct {
	http.ResponseWriter
	status    int
	bytes     int
	errorBody []byte
}

func (w *statusRecorder) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	if status >= http.StatusBadRequest {
		w.errorBody = make([]byte, 0, 512)
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusRecorder) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(data)
	w.bytes += n
	if w.status >= http.StatusBadRequest && len(w.errorBody) < 4096 {
		remaining := 4096 - len(w.errorBody)
		if len(data) > remaining {
			data = data[:remaining]
		}
		w.errorBody = append(w.errorBody, data...)
	}
	return n, err
}

func (w *statusRecorder) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func RequestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := newRequestID()
		r.Header.Set("X-Request-ID", requestID)
		w.Header().Set("X-Request-ID", requestID)

		startedAt := time.Now()
		recorder := &statusRecorder{ResponseWriter: w}
		panicRecovered := false
		defer func() {
			if recovered := recover(); recovered != nil {
				panicRecovered = true
				log.Printf("level=ERROR event=http_panic request_id=%q panic=%q stack=%q", requestID, fmt.Sprint(recovered), string(debug.Stack()))
				if recorder.status == 0 {
					response.WriteError(recorder, http.StatusInternalServerError, "internal server error")
				}
			}

			status := recorder.status
			if status == 0 {
				status = http.StatusOK
			}
			if panicRecovered {
				status = http.StatusInternalServerError
			}
			level := "INFO"
			if status >= http.StatusInternalServerError {
				level = "ERROR"
			} else if status >= http.StatusBadRequest {
				level = "WARN"
			}
			log.Printf("level=%s event=http_request request_id=%q method=%q path=%q status=%d bytes=%d duration_ms=%d remote_ip=%q response_message=%q",
				level, requestID, r.Method, uuidInPath.ReplaceAllString(r.URL.Path, "{id}"), status, recorder.bytes,
				time.Since(startedAt).Milliseconds(), remoteIP(r.RemoteAddr), errorMessage(recorder.errorBody))
		}()

		next.ServeHTTP(recorder, r)
	})
}

func newRequestID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err == nil {
		return hex.EncodeToString(value[:])
	}
	return fmt.Sprintf("%x-%x", time.Now().UnixNano(), requestSequence.Add(1))
}

func errorMessage(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err == nil {
		if message, ok := payload["message"].(string); ok {
			return message
		}
	}
	return strings.TrimSpace(string(body))
}
