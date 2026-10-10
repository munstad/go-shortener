package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

type createRequest struct {
	URL string `json:"url"`
}

type createResponse struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
}

// NewMux собирает маршруты сервиса.
func NewMux(store *Store) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /api/links", func(w http.ResponseWriter, r *http.Request) {
		var req createRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		if !validURL(req.URL) {
			writeError(w, http.StatusBadRequest, "url must be a valid http(s) address")
			return
		}
		l, err := store.Create(req.URL)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not create link")
			return
		}
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		writeJSON(w, http.StatusCreated, createResponse{
			Code:     l.Code,
			ShortURL: scheme + "://" + r.Host + "/" + l.Code,
		})
	})

	mux.HandleFunc("GET /api/links/{code}", func(w http.ResponseWriter, r *http.Request) {
		l, err := store.Get(r.PathValue("code"))
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "link not found")
			return
		}
		writeJSON(w, http.StatusOK, l)
	})

	mux.HandleFunc("GET /{code}", func(w http.ResponseWriter, r *http.Request) {
		l, err := store.Hit(r.PathValue("code"))
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "link not found")
			return
		}
		http.Redirect(w, r, l.URL, http.StatusFound)
	})

	return mux
}

func validURL(raw string) bool {
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return false
	}
	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
