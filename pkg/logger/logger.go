package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Logger struct {
	*slog.Logger
}

func New(ctx context.Context, cfg *Config) (*Logger, error) {
	if err := cfg.Validate(ctx); err != nil {
		return nil, err
	}

	opts := &slog.HandlerOptions{
		Level:     parseLevel(cfg.Level),
		AddSource: cfg.WithSource,
	}

	var handler slog.Handler
	switch strings.ToLower(cfg.Format) {
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	base := slog.New(handler).With("service", cfg.ServiceName)
	return &Logger{base}, nil
}

func (a *Logger) Print(v ...interface{}) {
	a.Logger.Info(fmt.Sprint(v...))
}

func (a *Logger) Printf(format string, v ...interface{}) {
	a.Logger.Info(fmt.Sprintf(format, v...))
}

func (a *Logger) Println(v ...interface{}) {
	a.Logger.Info(fmt.Sprintln(v...))
}
