package http

import (
	"encoding/json"
	"net/http"

	"github.com/marcelofabianov/redtogreen/internal/platform/adapter/web"
)

func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	logger := web.GetLogger(r.Context())
	logger.Info("identity context test handler called")

	response := map[string]string{
		"message": "users retrieved successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
