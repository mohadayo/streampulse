package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["status"] != "healthy" {
		t.Errorf("expected healthy status, got %v", resp["status"])
	}
	if resp["service"] != "event-processor" {
		t.Errorf("expected event-processor service, got %v", resp["service"])
	}
}

func TestProcessHandler(t *testing.T) {
	body := map[string]interface{}{
		"id":         "test-123",
		"event_type": "purchase",
		"payload":    map[string]interface{}{"amount": 42.0},
		"source":     "web",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/process", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	processHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["message"] != "Event processed" {
		t.Errorf("expected 'Event processed', got %v", resp["message"])
	}

	processed := resp["processed"].(map[string]interface{})
	labels := processed["labels"].(map[string]interface{})
	if labels["category"] != "conversion" {
		t.Errorf("expected 'conversion' category for purchase, got %v", labels["category"])
	}
	if labels["priority"] != "high" {
		t.Errorf("expected 'high' priority for purchase, got %v", labels["priority"])
	}
}

func TestProcessHandlerInvalidMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/process", nil)
	w := httptest.NewRecorder()
	processHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestProcessHandlerMissingEventType(t *testing.T) {
	body := map[string]interface{}{"payload": map[string]interface{}{}}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/process", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	processHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestProcessHandlerInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/process", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	processHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestClassifyEvent(t *testing.T) {
	tests := []struct {
		eventType string
		category  string
		priority  string
	}{
		{"page_view", "engagement", "low"},
		{"click", "interaction", "medium"},
		{"purchase", "conversion", "high"},
		{"error", "error", "critical"},
		{"unknown", "other", "low"},
	}

	for _, tt := range tests {
		labels := classifyEvent(tt.eventType)
		if labels["category"] != tt.category {
			t.Errorf("classifyEvent(%s) category = %s, want %s", tt.eventType, labels["category"], tt.category)
		}
		if labels["priority"] != tt.priority {
			t.Errorf("classifyEvent(%s) priority = %s, want %s", tt.eventType, labels["priority"], tt.priority)
		}
	}
}

func TestListProcessedHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/processed", nil)
	w := httptest.NewRecorder()
	listProcessedHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if _, ok := resp["processed_events"]; !ok {
		t.Error("expected processed_events in response")
	}
}
