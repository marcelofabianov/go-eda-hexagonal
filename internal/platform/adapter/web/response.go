package web

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/marcelofabianov/redtogreen/internal/platform/msg"
)

type SuccessResponse struct {
	Data any `json:"data"`
}

type ErrorResponse struct {
	Error ErrorDetails `json:"error"`
}

type ErrorDetails struct {
	Message string         `json:"message"`
	Code    string         `json:"code"`
	Context map[string]any `json:"context,omitempty"`
}

func Respond(w http.ResponseWriter, r *http.Request, code int, data any) {
	if code == http.StatusNoContent {
		w.WriteHeader(code)
		return
	}

	responsePayload := SuccessResponse{Data: data}
	jsonData, err := json.Marshal(responsePayload)
	if err != nil {
		logger := GetLogger(r.Context())
		logger.Error("failed to encode response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(jsonData)
}

func RespondError(w http.ResponseWriter, r *http.Request, err error) {
	logger := GetLogger(r.Context())
	logger.Error("api error", "error", err)

	var msgErr *msg.MessageError
	if !errors.As(err, &msgErr) {
		errorResponse := ErrorResponse{
			Error: ErrorDetails{
				Message: "An unexpected internal error occurred.",
				Code:    string(msg.CodeInternal),
			},
		}
		jsonData, _ := json.Marshal(errorResponse)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	publicError := msgErr.ToResponse()

	jsonData, jsonErr := json.Marshal(publicError)
	if jsonErr != nil {
		logger.Error("failed to encode error response", "error", jsonErr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(publicError.StatusCode)
	w.Write(jsonData)
}
