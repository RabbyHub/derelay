package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
)

func Logger() *zap.Logger {
	return logger.WithOptions(zap.AddCallerSkip(-1))
}

// Info ...
func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

// Warn ...
func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

// Error ...
func Error(msg string, err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	logger.Error(msg, fields...)
}

// Debug  ...
func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

// Fatal ...
func Fatal(msg string, err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	logger.Fatal(msg, fields...)
}

func ProductionModeWithoutStackTrace() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	config.DisableStacktrace = true
	config.OutputPaths = append(config.OutputPaths)

	buildLoggerWithConfig(config)
}
func buildLoggerWithConfig(config zap.Config) {
	zapLogger, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic("init zap logger: " + err.Error())
	}
	logger = zapLogger
}

func init() {
	ProductionModeWithoutStackTrace()
}

func Any(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}
