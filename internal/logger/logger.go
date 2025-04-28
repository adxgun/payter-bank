package logger

import (
	"context"
	"go.uber.org/zap"
)

var (
	baseLogger     *zap.Logger
	loggerFields   = "logger.fields"
	RequestFields  = "request.fields"
	AccountIDField = "account.id"
	FunctionName   = "function.name"
)

func init() {
	var err error
	baseLogger, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
}

// With adds structured fields to the logger in context and returns a new context.
func With(ctx context.Context, fields ...zap.Field) context.Context {
	data := ctx.Value(loggerFields)
	storedFields := make([]zap.Field, 0)
	if data != nil {
		storedFields = data.([]zap.Field)
	}

	storedFields = append(storedFields, fields...)
	return context.WithValue(ctx, loggerFields, storedFields)
}

// FromContext retrieves the structured fields from the context.
func FromContext(ctx context.Context) []zap.Field {
	data := ctx.Value(loggerFields)
	fields := make([]zap.Field, 0)
	if data != nil {
		fields = data.([]zap.Field)
	}

	return fields
}

func Info(ctx context.Context, message string, fields ...zap.Field) {
	storedFields := FromContext(ctx)
	storedFields = append(storedFields, fields...)
	baseLogger.Info(message, storedFields...)
}

func Error(ctx context.Context, message string, fields ...zap.Field) {
	storedFields := FromContext(ctx)
	storedFields = append(storedFields, fields...)
	baseLogger.Error(message, storedFields...)
}

func Warn(ctx context.Context, message string, fields ...zap.Field) {
	storedFields := FromContext(ctx)
	storedFields = append(storedFields, fields...)
	baseLogger.Warn(message, storedFields...)
}

func Fatal(ctx context.Context, message string, fields ...zap.Field) {
	storedFields := FromContext(ctx)
	storedFields = append(storedFields, fields...)
	baseLogger.Fatal(message, storedFields...)
}

// Base returns the base logger (without context).
func Base() *zap.Logger {
	return baseLogger
}
