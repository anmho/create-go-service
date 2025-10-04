package api

import (
	"log/slog"
	"net/http"

	"github.com/anmho/create-go-service/internal/posts"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func registerRoutes(
	r *chi.Mux,
	noteService *posts.Service,
) {

	r.Post("/notes", createNote(noteService))
	r.Get("/notes/{id}", getNote)
	r.Get("/notes", listNotes)
	r.Put("/notes/{id}", updateNote)
	r.Delete("/notes/{id}", deleteNote)

}

type CreateNoteRequest struct {
	Author    string `json:"author"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	IsPrivate bool   `json:"is_private"`
}

func createNote(noteService *posts.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := Body[CreateNoteRequest](r.Body)
		if err != nil {
			slog.Error("bad json body")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		id := uuid.New()

		author, err := uuid.Parse(params.Author)
		if err != nil {
			slog.Error("invalid author id")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		note, err := noteService.CreateNote(r.Context(), id, author, params.Content, params.Title)
		
		if err != nil {
			slog.Error("error creating note", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		
		w.WriteHeader(http.StatusCreated)
		JSON(w, note)
	}
}

func getNote(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("getting note"))
}

func listNotes(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("listing notes"))
}

func updateNote(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("updating note"))
}

func deleteNote(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("deleting note"))
}
