package httputil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSONResponse(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"message": "hello"}

	JSONResponse(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}

	var result map[string]string
	json.NewDecoder(w.Body).Decode(&result)
	if result["message"] != "hello" {
		t.Errorf("expected message 'hello', got '%s'", result["message"])
	}
}

func TestErrorResponse(t *testing.T) {
	w := httptest.NewRecorder()

	ErrorResponse(w, http.StatusNotFound, "not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	var result map[string]string
	json.NewDecoder(w.Body).Decode(&result)
	if result["error"] != "not found" {
		t.Errorf("expected error 'not found', got '%s'", result["error"])
	}
}
