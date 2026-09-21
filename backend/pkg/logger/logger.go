package logger

import (
	"log/slog"
	"os"
)

type Config struct {
	Environment string
}

func New(cfg Config) *slog.Logger {
	var handler slog.Handler

	switch cfg.Environment {
	case "production":
		handler = slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelInfo,
			},
		)
	case "dev":
		handler = slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					if a.Key == slog.TimeKey {
						a.Value = slog.StringValue(a.Value.Time().Format("2006-01-02 15:04:05"))
					}
					return a
				},
			},
		)
	default:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	return slog.New(handler)
}
