package grpc_zerolog_test

import (
	"context"
	"time"

	grpc_middleware "github.com/rkollar/go-grpc-middleware"
	grpc_zerolog "github.com/rkollar/go-grpc-middleware/logging/zerolog"
	"github.com/rkollar/go-grpc-middleware/logging/zerolog/ctxzerolog"
	grpc_ctxtags "github.com/rkollar/go-grpc-middleware/tags"
	pb_testproto "github.com/rkollar/go-grpc-middleware/testing/testproto"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

var (
	zerologLogger zerolog.Logger
	customFunc    grpc_zerolog.CodeToLevel
)

// Initialization shows a relatively complex initialization sequence.
func Example_initialization() {
	// Shared options for the logger, with a custom gRPC code to log level function.
	opts := []grpc_zerolog.Option{
		grpc_zerolog.WithLevels(customFunc),
	}
	// Make sure that log statements internal to gRPC library are logged using the zerologLogger as well.
	grpc_zerolog.ReplaceGrpcLoggerV2(zerologLogger)
	// Create a server, make sure we put the grpc_ctxtags context before everything else.
	_ = grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_zerolog.UnaryServerInterceptor(zerologLogger, opts...),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_zerolog.StreamServerInterceptor(zerologLogger, opts...),
		),
	)
}

// Initialization shows an initialization sequence with the duration field generation overridden.
func Example_initializationWithDurationFieldOverride() {
	opts := []grpc_zerolog.Option{
		grpc_zerolog.WithDurationField(func(duration time.Duration) map[string]interface{} {
			return map[string]interface{}{"grpc.time_ns": duration.Nanoseconds()}
		}),
	}

	_ = grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_zerolog.UnaryServerInterceptor(zerologLogger, opts...),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_zerolog.StreamServerInterceptor(zerologLogger, opts...),
		),
	)
}

// Simple unary handler that adds custom fields to the requests's context. These will be used for all log statements.
func ExampleExtract_unary() {
	_ = func(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
		// Add fields the ctxtags of the request which will be added to all extracted loggers.
		grpc_ctxtags.Extract(ctx).Set("custom_tags.string", "something").Set("custom_tags.int", 1337)

		// Extract a single request-scoped zerolog.Logger and log messages. (containing the grpc.xxx tags)
		l := ctxzerolog.Extract(ctx)
		l.Info().Msg("some ping")
		l.Info().Msg("another ping")
		return &pb_testproto.PingResponse{Value: ping.Value}, nil
	}
}

func Example_initializationWithDecider() {
	opts := []grpc_zerolog.Option{
		grpc_zerolog.WithDecider(func(fullMethodName string, err error) bool {
			// will not log gRPC calls if it was a call to healthcheck and no error was raised
			if err == nil && fullMethodName == "foo.bar.healthcheck" {
				return false
			}

			// by default everything will be logged
			return true
		}),
	}

	_ = []grpc.ServerOption{
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_zerolog.StreamServerInterceptor(zerolog.Nop(), opts...)),
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_zerolog.UnaryServerInterceptor(zerolog.Nop(), opts...)),
	}
}
