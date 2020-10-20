package grpc_zerolog

import (
	"context"

	"github.com/rkollar/go-grpc-middleware/logging/zerolog/ctxzerolog"
	"github.com/rs/zerolog"
)

// AddFields adds zerolog fields to the logger.
// Deprecated: should use the ctxzerolog.AddFields instead
func AddFields(ctx context.Context, fields map[string]interface{}) {
	ctxzerolog.AddFields(ctx, fields)
}

// Extract takes the call-scoped Logger from grpc_zerolog middleware.
// Deprecated: should use the ctxzerolog.Extract instead
func Extract(ctx context.Context) zerolog.Logger {
	return ctxzerolog.Extract(ctx)
}
