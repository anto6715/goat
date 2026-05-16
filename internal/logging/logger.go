package logging

import (
	"log/slog"
	"os"
	"strings"
)

func New() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				return slog.String("time", a.Value.Time().Format("2006-01-02T15:04:05"))
			case slog.LevelKey:
				return slog.String("level", strings.ToLower(a.Value.String()))
			default:
				return a
			}
		},
	}

	return slog.New(slog.NewTextHandler(os.Stderr, opts))
}
