package response

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
)

type Envelope struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteSuccess(w http.ResponseWriter, status int, message string, data, meta any) {
	WriteJSON(w, status, Envelope{Success: true, Message: message, Data: data, Meta: meta})
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, Envelope{Success: false, Message: message})
}
func WriteInternalError(w http.ResponseWriter, r *http.Request, message string, err error) {
	WriteInternalErrorStatus(w, r, http.StatusInternalServerError, message, err)
}

func WriteInternalErrorStatus(w http.ResponseWriter, r *http.Request, status int, message string, err error) {
	requestID := r.Header.Get("X-Request-ID")
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		log.Printf("level=ERROR event=internal_error request_id=%q error=%q sqlstate=%q position=%d",
			requestID, err, pgErr.Code, pgErr.Position)
	} else {
		log.Printf("level=ERROR event=internal_error request_id=%q error=%q", requestID, err)
	}
	WriteError(w, status, message)
}
