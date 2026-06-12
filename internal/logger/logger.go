package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log defaults to a no-op logger so the package-level helpers are safe to call
// before InitLogger runs (e.g. in tests or early startup); InitLogger swaps in
// the configured logger.
var Log = zap.NewNop()

// InitLogger builds the global logger; an environment with the "prod" prefix
// gets JSON output with ISO8601 timestamps, anything else gets colored,
// human-readable development output.
func InitLogger(environment string) error {
	var config zap.Config

	if strings.HasPrefix(environment, "prod") {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	logger, err := config.Build(
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return err
	}

	Log = logger
	return nil
}

// Sync flushes any buffered log entries; call it before exit (errors ignored,
// as stdout/stderr sync can fail harmlessly on some platforms).
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	Log.Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Log.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Log.Error(msg, fields...)
}

// Fatal logs at fatal level and then calls os.Exit(1).
func Fatal(msg string, fields ...zap.Field) {
	Log.Fatal(msg, fields...)
}

func With(fields ...zap.Field) *zap.Logger {
	return Log.With(fields...)
}
