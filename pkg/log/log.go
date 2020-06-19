package log

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"os"
)

func init() {
	SetGlobalOut(os.Stdout)
}

func SetGlobalOut(w io.Writer) {
	log.Logger = zerolog.New(w).With().Timestamp().Logger()
}

var errorLogger = zerolog.New(os.Stderr).With().Caller().Timestamp().Logger()

func SetGlobalErr(w io.Writer) {
	errorLogger = zerolog.New(w).With().Caller().Timestamp().Logger()
}

func Default() *zerolog.Event {
	return log.Info().Str("severity", "DEFAULT")
}

func Debug(ctx context.Context) *zerolog.Event {
	return WithTrace(ctx, log.Debug()).Str("severity", "DEBUG")
}

func Info(ctx context.Context) *zerolog.Event {
	return WithTrace(ctx, log.Info()).Str("severity", "INFO")
}

func Notice(ctx context.Context) *zerolog.Event {
	return WithTrace(ctx, log.Info()).Str("severity", "NOTICE")
}

func Warn(ctx context.Context, skip ...int) *zerolog.Event {
	logger := skipLogger(log.Logger, skip...)
	return WithTrace(ctx, logger.Warn()).Str("severity", "WARNING")
}

func Error(ctx context.Context, skip ...int) *zerolog.Event {
	logger := skipLogger(errorLogger, skip...)
	return WithTrace(ctx, logger.Error()).Str("severity", "ERROR").Str("@type", ErrorReport)
}

func Fatal(ctx context.Context, skip ...int) *zerolog.Event {
	logger := skipLogger(errorLogger, skip...)
	return WithTrace(ctx, logger.Error()).Str("severity", "CRITICAL").Str("@type", ErrorReport)
}

func Critical(ctx context.Context, skip ...int) *zerolog.Event {
	logger := skipLogger(errorLogger, skip...)
	return WithTrace(ctx, logger.Error()).Str("severity", "CRITICAL").Str("@type", ErrorReport)
}

func Alert(ctx context.Context, skip ...int) *zerolog.Event {
	logger := skipLogger(errorLogger, skip...)
	return WithTrace(ctx, logger.Error()).Str("severity", "ALERT").Str("@type", ErrorReport)
}

func Emergency(ctx context.Context, skip ...int) *zerolog.Event {
	logger := skipLogger(errorLogger, skip...)
	return WithTrace(ctx, logger.Error()).Str("severity", "EMERGENCY").Str("@type", ErrorReport)
}

func skipLogger(logger zerolog.Logger, skip ...int) zerolog.Logger {
	if skip != nil && len(skip) > 0 {
		skipCount := zerolog.CallerSkipFrameCount + skip[0]
		logger = zerolog.New(os.Stderr).With().CallerWithSkipFrameCount(skipCount).Timestamp().Logger()
	}
	return logger
}
