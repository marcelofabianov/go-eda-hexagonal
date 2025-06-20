package http

import (
	"github.com/go-chi/chi/v5"
)

type Router struct {
	*chi.Mux
}

func NewIdentityRouter(
	createUserHandler *CreateUserHandler,
) *Router {
	r := chi.NewRouter()

	r.Route("/users", func(r chi.Router) {
		r.Post("/", createUserHandler.Handle)
		// r.Get("/", getUsersHandler.Handle)
		// r.Get("/{userID}", getUserByIDHandler.Handle)
		// r.Put("/{userID}", updateUserHandler.Handle)
		// r.Delete("/{userID}", deleteUserHandler.Handle)
		// r.Patch("/{userID}/archive", archiveUserHandler.Handle)
		// r.Patch("/{userID}/unarchive", unarchiveUserHandler.Handle)
	})

	return &Router{Mux: r}
}
