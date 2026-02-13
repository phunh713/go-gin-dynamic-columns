package config

import (
	"log/slog"
	"os"
)

type LogPayload map[string]any

const LogPayloadKey = "log_payload"

func NewLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	)
}
