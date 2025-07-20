package logger

import (
	"log/slog"
	"strings"
)

const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

func parseLevel(l string) slog.Level {
	levelMap := map[string]slog.Level{
		LevelDebug: slog.LevelDebug,
		LevelInfo:  slog.LevelInfo,
		LevelWarn:  slog.LevelWarn,
		LevelError: slog.LevelError,
	}

	if level, ok := levelMap[strings.ToLower(l)]; ok {
		return level
	}
	return slog.LevelInfo
}

func getLevelMap() map[string]slog.Level {
	return map[string]slog.Level{
		LevelDebug: slog.LevelDebug,
		LevelInfo:  slog.LevelInfo,
		LevelWarn:  slog.LevelWarn,
		LevelError: slog.LevelError,
	}
}
