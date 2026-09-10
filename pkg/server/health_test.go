package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthRoutes(t *testing.T) {
	srv := NewServer()
	registerHealthRoutes(srv.mux)

	paths := []string{healthPath, healthzPath, readyzPath}
	for _, path := range paths {
		for _, method := range []string{http.MethodGet, http.MethodHead} {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(method, path, nil)
			srv.mux.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("%s %s: got %d, want %d", method, path, rec.Code, http.StatusOK)
			}
		}
	}
}
