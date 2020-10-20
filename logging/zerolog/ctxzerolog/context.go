package ctxzerolog

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/rs/zerolog"
)

type ctxMarker struct{}

type ctxLogger struct {
	logger zerolog.Logger
	fields map[string]interface{}
}

var (
	ctxMarkerKey = &ctxMarker{}
	nullLogger   = zerolog.Nop()
)

// AddFields adds fields to the logger.
func AddFields(ctx context.Context, fields map[string]interface{}) {
	l, ok := ctx.Value(ctxMarkerKey).(*ctxLogger)
	if !ok || l == nil {
		return
	}
	for k, v := range fields {
		l.fields[k] = v
	}
}

// Extract takes the call-scoped Logger from grpc_zerolog middleware.
//
// It always returns a Logger that has all the grpc_ctxtags updated.
func Extract(ctx context.Context) zerolog.Logger {
	l, ok := ctx.Value(ctxMarkerKey).(*ctxLogger)
	if !ok || l == nil {
		return nullLogger
	}
	// Add grpc_ctxtags tags metadata until now.
	fields := TagsToFields(ctx)
	// Add fields added until now.
	for k, v := range l.fields {
		fields[k] = v
	}
	return l.logger.With().Fields(fields).Logger()
}

// TagsToFields transforms the Tags on the supplied context into fields.
func TagsToFields(ctx context.Context) map[string]interface{} {
	fields := make(map[string]interface{})
	tags := grpc_ctxtags.Extract(ctx)
	for k, v := range tags.Values() {
		fields[k] = v
	}
	return fields
}

// ToContext adds the zerolog.Logger to the context for extraction later.
// Returning the new context that has been created.
func ToContext(ctx context.Context, logger zerolog.Logger) context.Context {
	l := &ctxLogger{
		logger: logger,
	}
	return context.WithValue(ctx, ctxMarkerKey, l)
}
