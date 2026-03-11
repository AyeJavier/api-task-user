package handler

import (
	"encoding/json"
	"net/http"

	"github.com/javier/api-task-user/internal/adapter/inbound/http/dto"
)

// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// respondError writes a standardized JSON error response.
func respondError(w http.ResponseWriter, status int, message string, detail any) {
	respondJSON(w, status, dto.ErrorResponse{
		Error:  message,
		Detail: detail,
	})
}
