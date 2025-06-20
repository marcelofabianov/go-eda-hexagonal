package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/marcelofabianov/redtogreen/internal/platform/adapter/web"
)

func UnarchiveUserHandler(w http.ResponseWriter, r *http.Request) {
	logger := web.GetLogger(r.Context())

	userID := chi.URLParam(r, "userID")

	logger.Info("unarchive user request received", "user_id", userID)

	response := map[string]string{
		"status":  "user unarchived successfully",
		"user_id": userID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
