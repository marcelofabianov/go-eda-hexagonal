package http

import (
	"encoding/json"
	"net/http"

	"github.com/marcelofabianov/redtogreen/internal/contexts/identity/domain/user"
	"github.com/marcelofabianov/redtogreen/internal/platform/adapter/validator"
	"github.com/marcelofabianov/redtogreen/internal/platform/adapter/web"
	"github.com/marcelofabianov/redtogreen/internal/platform/types"
)

type CreateUserRequest struct {
	Name                 string          `json:"name" validate:"required,min=2,max=100"`
	Email                string          `json:"email" validate:"required,email"`
	Password             string          `json:"password" validate:"required,min=10,max=72"`
	PasswordConfirmation string          `json:"password_confirmation" validate:"required,eqfield=Password"`
	Phone                string          `json:"phone" validate:"required,min=10,max=30"`
	Preferences          json.RawMessage `json:"preferences,omitempty"`
}

type CreateUserHandler struct {
	command   user.CreateUserCommand
	validator *validator.Validator
}

func NewCreateUserHandler(cmd user.CreateUserCommand, v *validator.Validator) *CreateUserHandler {
	return &CreateUserHandler{
		command:   cmd,
		validator: v,
	}
}

func (h *CreateUserHandler) Handle(w http.ResponseWriter, r *http.Request) {
	logger := web.GetLogger(r.Context())

	var req CreateUserRequest
	if err := web.Decode(w, r, &req); err != nil {
		logger.Error("failed to decode request body", "error", err)
		web.RespondError(w, r, err)
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		logger.Error("request validation failed", "error", err)
		web.RespondError(w, r, err)
		return
	}

	traceID := web.GetTraceID(r.Context())
	userAuthorID := web.GetUserAuthorID(r.Context())
	correlationID := web.GetCorrelationID(r.Context())

	var previousEventID types.NullableUUID
	var causationID types.NullableUUID

	newUserUseCaseInput := user.NewUserInput{
		Name:        req.Name,
		Email:       req.Email,
		Phone:       req.Phone,
		Password:    req.Password,
		Preferences: req.Preferences,
	}

	commandInput := user.CreateUserCommandInput{
		CorrelationID:   correlationID,
		TraceID:         traceID,
		UserAuthorID:    userAuthorID,
		PreviousEventID: previousEventID,
		CausationID:     causationID,
		NewUserInput:    newUserUseCaseInput,
	}

	output, err := h.command.Execute(r.Context(), commandInput)
	if err != nil {
		logger.Error("failed to execute create user command", "error", err)
		web.RespondError(w, r, err)
		return
	}

	web.Respond(w, r, http.StatusCreated, output.User)
}
