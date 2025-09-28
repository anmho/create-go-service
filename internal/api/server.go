package api

import (
	"net/http"

	"github.com/anmho/create-go-service/internal/notes"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func New(noteService *notes.Service) *chi.Mux {
	r := chi.NewMux()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	registerRoutes(r, noteService)

	return r
}
