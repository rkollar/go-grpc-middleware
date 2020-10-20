package ctxzerolog_test

import (
	"context"

	"github.com/rkollar/go-grpc-middleware/logging/zerolog/ctxzerolog"
	"github.com/rkollar/go-grpc-middleware/tags"
	pb_testproto "github.com/rkollar/go-grpc-middleware/testing/testproto"
	"github.com/rs/zerolog"
)

var zerologLogger zerolog.Logger

// Simple unary handler that adds custom fields to the requests's context. These will be used for all log statements.
func ExampleExtract_unary() {
	_ = func(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
		// Add fields the ctxtags of the request which will be added to all extracted loggers.
		grpc_ctxtags.Extract(ctx).Set("custom_tags.string", "something").Set("custom_tags.int", 1337)

		// Extract a single request-scoped zerolog.Logger and log messages.
		l := ctxzerolog.Extract(ctx)
		l.Info().Msgf("some ping")
		l.Info().Msgf("another ping")
		return &pb_testproto.PingResponse{Value: ping.Value}, nil
	}
}
