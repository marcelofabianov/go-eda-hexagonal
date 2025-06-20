package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/marcelofabianov/redtogreen/internal/platform/adapter/web"
)

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	logger := web.GetLogger(r.Context())

	userID := chi.URLParam(r, "userID")

	logger.Info("delete user request received", "user_id", userID)

	w.WriteHeader(http.StatusNoContent)
}
