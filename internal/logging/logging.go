package logging

import (
	"flag"
	"log/slog"
	"os"
)

func init() {
	enableDebug := flag.Bool("d", false, "enable DEBUG output")
	flag.Parse()

	logLevel := slog.LevelInfo
	if *enableDebug {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	slog.SetDefault(logger)
	slog.Debug("Debug logging enabled")
}
