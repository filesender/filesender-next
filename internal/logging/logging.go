package logging

import (
	"log/slog"
	"os"
)

func init() {
	SetLogLevel(slog.LevelInfo)
}

func SetLogLevel(logLevel slog.Leveler) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	slog.SetDefault(logger)
	slog.Debug("Debug logging enabled")
}
