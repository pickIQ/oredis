package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pickIQ/oredis/internal/response"
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

func healtzHandler(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "alive"})
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"version": "0.1.0"})

}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	value, found := h.store.Get(key)

	if found {
		response.JSON(w, http.StatusOK, valueResponse{Value: value})
	} else {
		response.Error(w, http.StatusNotFound, "key not found")
	}
}

func (h *Handler) Set(w http.ResponseWriter, r *http.Request) {
	var req setRequest
	key := r.PathValue("key")

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid input data")
		return
	}
	if err := h.store.Set(key, req.Value, time.Duration(req.TTLSeconds)*time.Second); err != nil {
		response.Error(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	result := h.store.Delete(key)

	if !result {
		response.Error(w, http.StatusNotFound, "key not found or expired")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func panicHandler(w http.ResponseWriter, r *http.Request) {
	panic("тестовая паника")
}
func NewRouter(s store.Store) http.Handler {
	h := NewHandler(s)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /keys/{key}", h.Get)
	mux.HandleFunc("PUT /keys/{key}", h.Set)
	mux.HandleFunc("DELETE /keys/{key}", h.Delete)
	mux.HandleFunc("/healtz", healtzHandler)
	mux.HandleFunc("/version", versionHandler)
	mux.HandleFunc("/panic", panicHandler)

	return mux
}
