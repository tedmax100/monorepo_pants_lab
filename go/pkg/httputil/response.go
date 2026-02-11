// Package httputil provides shared HTTP response helpers for all Go services.
package httputil

import (
	"encoding/json"
	"net/http"
)

// JSONResponse writes a JSON response with the given status code.
func JSONResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// ErrorResponse writes a JSON error response.
func ErrorResponse(w http.ResponseWriter, status int, message string) {
	JSONResponse(w, status, map[string]string{"error": message})
}
