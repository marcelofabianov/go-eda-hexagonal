package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/trace"

	"github.com/marcelofabianov/redtogreen/internal/platform/msg"
	"github.com/marcelofabianov/redtogreen/internal/platform/types"
)

const (
	UserAuthorIDCtxKey  contextKey = "userAuthorID"
	CorrelationIDCtxKey contextKey = "correlationID"
)

func Decode(w http.ResponseWriter, r *http.Request, val any) error {
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodySize)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(val); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			message := fmt.Sprintf("Request body contains badly-formed JSON at character %d.", syntaxError.Offset)
			return msg.NewValidationError(err,
				map[string]any{"offset": syntaxError.Offset},
				message,
			)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return msg.NewValidationError(err, nil, "Request body contains badly-formed JSON.")

		case errors.As(err, &unmarshalTypeError):
			context := map[string]any{"offset": unmarshalTypeError.Offset}
			message := "Request body contains an incorrect JSON type."
			if unmarshalTypeError.Field != "" {
				context["field"] = unmarshalTypeError.Field
				context["type"] = unmarshalTypeError.Type.String()
				message = fmt.Sprintf("The field '%s' contains an incorrect JSON type.", unmarshalTypeError.Field)
			}
			return msg.NewValidationError(err, context, message)

		case errors.Is(err, io.EOF):
			return msg.NewValidationError(err, nil, "Request body must not be empty.")

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return msg.NewInternalError(err, nil)
		}
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return msg.NewValidationError(errors.New("body must only contain a single JSON value"), nil, "")
	}

	return nil
}

func GetTraceID(ctx context.Context) types.UUID {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		uuid, err := types.ParseUUID(spanCtx.TraceID().String())
		if err == nil {
			return uuid
		}
	}

	reqID := middleware.GetReqID(ctx)
	if reqID != "" {
		uuid, err := types.ParseUUID(reqID)
		if err == nil {
			return uuid
		}
	}

	newID, _ := types.NewUUID()
	return newID
}

func GetUserAuthorID(ctx context.Context) types.NullableUUID {
	if id, ok := ctx.Value(UserAuthorIDCtxKey).(types.NullableUUID); ok {
		return id
	}
	return types.NewNullUUID()
}

func GetCorrelationID(ctx context.Context) types.UUID {
	if id, ok := ctx.Value(CorrelationIDCtxKey).(types.UUID); ok {
		return id
	}
	newID, _ := types.NewUUID()
	return newID
}
