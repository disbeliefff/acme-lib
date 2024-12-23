package logging

import "go.uber.org/zap"

type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

type ZapLogger struct {
	logger *zap.Logger
}

func NewZapLogger(logger *zap.Logger) *ZapLogger {
	return &ZapLogger{logger: logger}
}

func (z *ZapLogger) Info(msg string, fields ...interface{}) {
	zapFields := make([]zap.Field, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			continue
		}
		zapFields[i/2] = zap.Any(key, fields[i+1])
	}
	z.logger.Info(msg, zapFields...)
}

func (z *ZapLogger) Error(msg string, fields ...interface{}) {
	zapFields := make([]zap.Field, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			continue
		}
		zapFields[i/2] = zap.Any(key, fields[i+1])
	}
	z.logger.Error(msg, zapFields...)
}

func (z *ZapLogger) Debug(msg string, fields ...interface{}) {
	zapFields := make([]zap.Field, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			continue
		}
		zapFields[i/2] = zap.Any(key, fields[i+1])
	}
	z.logger.Debug(msg, zapFields...)
}
