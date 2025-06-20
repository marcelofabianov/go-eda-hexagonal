package logger

import (
	"log/slog"

	"github.com/marcelofabianov/redtogreen/internal/platform/msg"
	"github.com/marcelofabianov/redtogreen/internal/platform/types"
)

func TraceID(id string) slog.Attr {
	return slog.String("trace_id", id)
}

func EventID(id types.UUID) slog.Attr {
	return slog.String("event_id", id.String())
}

func UserID(id types.NullableUUID) slog.Attr {
	if id.IsValid() {
		actualUUID, _ := id.GetUUID()
		return slog.String("user_id", actualUUID.String())
	}
	return slog.Attr{}
}

func Component(name string) slog.Attr {
	return slog.String("component", name)
}

func Action(name string) slog.Attr {
	return slog.String("action", name)
}

func Err(err error) slog.Attr {
	return slog.String("err", err.Error())
}

func EventType(eventType string) slog.Attr {
	return slog.String("event_type", eventType)
}

func ErrorCode(code msg.ErrorCode) slog.Attr {
	return slog.String("error_code", string(code))
}
