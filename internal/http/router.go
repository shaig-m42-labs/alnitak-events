package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/m42-labs/alnitak-events/internal/service"
)

func Router(svc *service.Service) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"service": "alnitak-events", "status": "UP"})
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte("orion_service_up 1\n"))
	})
	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		events, err := svc.Events(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, events)
	})
	mux.HandleFunc("/deliveries", func(w http.ResponseWriter, r *http.Request) {
		deliveries, err := svc.Deliveries(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, deliveries)
	})
	return correlation(mux)
}

func correlation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Correlation-Id")
		if id == "" {
			id = time.Now().UTC().Format("20060102150405.000000000")
		}
		w.Header().Set("X-Correlation-Id", id)
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
