package utils

import (
	"context"
	"log/slog"
	"os"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
)

func LogFatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}

func LogFor(ctx context.Context) *slog.Logger {
	if val := ctx.Value(common.CtxLogger); val != nil {
		if logger, ok := val.(*slog.Logger); ok {
			return logger
		}
	}
	return slog.Default()
}
