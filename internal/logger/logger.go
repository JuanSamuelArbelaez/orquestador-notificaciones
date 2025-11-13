// logger/logger.go
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

// New crea una nueva instancia de logger estructurado
func New(serviceName string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()

	// Ajustes del encoder (estructura del JSON)
	cfg.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Salida estándar (puede ser capturada por Loki, Docker, etc.)
	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stderr"}

	// Construye el logger
	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	// Agrega el nombre del servicio como campo fijo
	logger = logger.With(zap.String("service", serviceName))

	return logger, nil
}
