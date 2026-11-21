package response

import (
	"encoding/json"
	"net/http"
)

type Envelope struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, payload Envelope) {
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
