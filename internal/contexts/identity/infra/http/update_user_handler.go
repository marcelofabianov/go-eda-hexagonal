package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/marcelofabianov/redtogreen/internal/platform/adapter/web"
)

type UpdateUserRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	logger := web.GetLogger(r.Context())

	userID := chi.URLParam(r, "userID")

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request body for update", "error", err, "user_id", userID)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	logger.Info("update user request received", "user_id", userID, "new_name", req.Name)

	response := map[string]interface{}{
		"status":       "user updated successfully",
		"user_id":      userID,
		"updated_data": req,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
