package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUserJSON(t *testing.T) {
	user := User{
		ID:        "u-1",
		Name:      "Alice",
		Email:     "alice@example.com",
		CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("failed to marshal user: %v", err)
	}

	var decoded User
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal user: %v", err)
	}

	if decoded.Name != "Alice" {
		t.Errorf("expected name 'Alice', got '%s'", decoded.Name)
	}
}

func TestOrderJSON(t *testing.T) {
	order := Order{
		ID:        "o-1",
		UserID:    "u-1",
		Product:   "Widget",
		Quantity:  3,
		Total:     29.97,
		CreatedAt: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(order)
	if err != nil {
		t.Fatalf("failed to marshal order: %v", err)
	}

	var decoded Order
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal order: %v", err)
	}

	if decoded.Total != 29.97 {
		t.Errorf("expected total 29.97, got %f", decoded.Total)
	}
}
