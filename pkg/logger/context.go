package logger

import (
	"context"
	"github.com/google/uuid"
)

type contextKey string

const (
	// CorrelationIDKey is the key for correlation ID in context
	CorrelationIDKey contextKey = "correlation_id"
	// CorrelationIDHeader is the HTTP header name for correlation ID
	CorrelationIDHeader = "X-Correlation-ID"
)

// WithCorrelationID adds correlation ID to context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

// GetCorrelationID retrieves correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return id
	}
	return ""
}

// GenerateCorrelationID generates a new correlation ID
func GenerateCorrelationID() string {
	return uuid.New().String()
}

// EnsureCorrelationID returns existing correlation ID or generates new one
func EnsureCorrelationID(ctx context.Context) (context.Context, string) {
	id := GetCorrelationID(ctx)
	if id == "" {
		id = GenerateCorrelationID()
		ctx = WithCorrelationID(ctx, id)
	}
	return ctx, id
}
