package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Event struct {
	ID        string                 `json:"id"`
	EventType string                 `json:"event_type"`
	Payload   map[string]interface{} `json:"payload"`
	Source    string                  `json:"source"`
	Timestamp string                 `json:"timestamp"`
}

type ProcessedEvent struct {
	Event
	ProcessedAt string            `json:"processed_at"`
	Labels      map[string]string `json:"labels"`
}

type ProcessorStore struct {
	mu     sync.RWMutex
	events []ProcessedEvent
}

var store = &ProcessorStore{}

func (s *ProcessorStore) Add(e ProcessedEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, e)
}

func (s *ProcessorStore) GetAll(limit int) []ProcessedEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 || limit > len(s.events) {
		limit = len(s.events)
	}
	start := len(s.events) - limit
	if start < 0 {
		start = 0
	}
	result := make([]ProcessedEvent, limit)
	copy(result, s.events[start:])
	return result
}

func (s *ProcessorStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.events)
}

func classifyEvent(eventType string) map[string]string {
	labels := map[string]string{"category": "other", "priority": "low"}

	switch eventType {
	case "page_view", "scroll":
		labels["category"] = "engagement"
		labels["priority"] = "low"
	case "click", "form_submit":
		labels["category"] = "interaction"
		labels["priority"] = "medium"
	case "purchase", "signup":
		labels["category"] = "conversion"
		labels["priority"] = "high"
	case "error", "crash":
		labels["category"] = "error"
		labels["priority"] = "critical"
	}
	return labels
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"service":   "event-processor",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func processHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var event Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		log.Printf("[WARN] Invalid event payload: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON payload"})
		return
	}

	if event.EventType == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "event_type is required"})
		return
	}

	processed := ProcessedEvent{
		Event:       event,
		ProcessedAt: time.Now().UTC().Format(time.RFC3339),
		Labels:      classifyEvent(event.EventType),
	}

	store.Add(processed)
	log.Printf("[INFO] Processed event %s (type=%s, category=%s, priority=%s)",
		event.ID, event.EventType, processed.Labels["category"], processed.Labels["priority"])

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Event processed",
		"processed": processed,
	})
}

func listProcessedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	events := store.GetAll(100)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"processed_events": events,
		"total":            store.Count(),
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/process", processHandler)
	mux.HandleFunc("/processed", listProcessedHandler)

	addr := fmt.Sprintf("0.0.0.0:%s", port)
	log.Printf("[INFO] Starting event-processor on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("[FATAL] Server failed: %v", err)
	}
}
