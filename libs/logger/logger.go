package logger

import (
	"log/slog"
	"os"
)

// Setup configures the global slog default logger.
// env="development" uses a human-readable text handler at DEBUG level.
// Any other value uses a structured JSON handler at INFO level.
func Setup(env string) {
	var handler slog.Handler
	if env == "development" {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}
	slog.SetDefault(slog.New(handler))
}

// Sync is a no-op kept for API compatibility.
// slog writes synchronously; no flush is needed.
func Sync() {}
