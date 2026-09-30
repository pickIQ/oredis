package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pickIQ/oredis/internal/store"
)

type valueResponse struct {
	Value string `json:"value"`
}

type setRequest struct {
	Value      string `json:"value"`
	TTLSeconds int    `json:"ttl_seconds"`
}

type Handler struct {
	store store.Store
}

func NewHandler(s store.Store) *Handler {
	return &Handler{
		store: s,
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "alive"})
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"version": "0.1.0"})

}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	value, found := h.store.Get(key)

	if found {
		writeJSON(w, http.StatusOK, valueResponse{Value: value})
	} else {
		writeError(w, http.StatusNotFound, "key not found")
	}
}

func (h *Handler) Set(w http.ResponseWriter, r *http.Request) {
	var req setRequest
	key := r.PathValue("key")

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid input data")
		return
	}
	if err := h.store.Set(key, req.Value, time.Duration(req.TTLSeconds)*time.Second); err != nil {
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	result := h.store.Delete(key)

	if !result {
		writeError(w, http.StatusNotFound, "key not found or expired")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func NewRouter(s store.Store) http.Handler {
	h := NewHandler(s)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /keys/{key}", h.Get)
	mux.HandleFunc("PUT /keys/{key}", h.Set)
	mux.HandleFunc("DELETE /keys/{key}", h.Delete)
	mux.HandleFunc("/healtz", healthzHandler)
	mux.HandleFunc("/version", versionHandler)

	return mux
}
