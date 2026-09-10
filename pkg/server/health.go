package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

const (
	healthPath  = "/health"
	healthzPath = "/healthz"
	readyzPath  = "/readyz"
)

func registerHealthRoutes(mux *chi.Mux) {
	ok := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ok"))
	}

	// /health se mantiene por compat. Helm y el chart usan /healthz y /readyz.
	mux.Get(healthPath, ok)
	mux.Get(healthzPath, ok)
	mux.Get(readyzPath, ok)
	mux.Head(healthPath, ok)
	mux.Head(healthzPath, ok)
	mux.Head(readyzPath, ok)
}
