package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestWithAndFromContext(t *testing.T) {
	ctx := context.Background()

	field1 := zap.String("key1", "value1")
	field2 := zap.Int("key2", 2)

	ctxWithFields := With(ctx, field1)
	ctxWithMoreFields := With(ctxWithFields, field2)

	fields := FromContext(ctxWithMoreFields)

	assert.Contains(t, fields, field1)
	assert.Contains(t, fields, field2)
	assert.Len(t, fields, 2)
}

func TestBaseLoggerIsNotNil(t *testing.T) {
	assert.NotNil(t, Base())
}

func TestLoggingWithZapTestLogger(t *testing.T) {
	tLogger := zaptest.NewLogger(t)
	defer tLogger.Sync()

	ctx := context.Background()
	ctx = With(ctx, zap.String("user", "hammed"))

	// Just ensure these don't panic
	Info(ctx, "info log")
	Error(ctx, "error log")
	Warn(ctx, "warn log")
}
