// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_zerolog

import (
	"context"
	"path"
	"time"

	"github.com/rkollar/go-grpc-middleware/logging/zerolog/ctxzerolog"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

var (
	// ClientField is used in every client-side log statement made through grpc_zerolog. Can be overwritten before initialization.
	ClientField = func(c zerolog.Context) zerolog.Context {
		return c.Str("span.kind", "client")
	}
)

// UnaryClientInterceptor returns a new unary client interceptor that optionally logs the execution of external gRPC calls.
func UnaryClientInterceptor(logger zerolog.Logger, opts ...Option) grpc.UnaryClientInterceptor {
	o := evaluateClientOpt(opts)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		logger := newClientLogger(ctx, method, logger)
		startTime := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		newCtx := ctxzerolog.ToContext(ctx, logger)
		logFinalClientLine(newCtx, o, startTime, err, "finished client unary call")
		return err
	}
}

// StreamClientInterceptor returns a new streaming client interceptor that optionally logs the execution of external gRPC calls.
func StreamClientInterceptor(logger zerolog.Logger, opts ...Option) grpc.StreamClientInterceptor {
	o := evaluateClientOpt(opts)
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		logger := newClientLogger(ctx, method, logger)
		startTime := time.Now()
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		newCtx := ctxzerolog.ToContext(ctx, logger)
		logFinalClientLine(newCtx, o, startTime, err, "finished client streaming call")
		return clientStream, err
	}
}

func logFinalClientLine(ctx context.Context, o *options, startTime time.Time, err error, msg string) {
	code := o.codeFunc(err)
	level := o.levelFunc(code)
	duration := o.durationFunc(time.Now().Sub(startTime))
	o.messageFunc(ctx, msg, level, code, err, duration)
}

func newClientLogger(ctx context.Context, fullMethodString string, logger zerolog.Logger) zerolog.Logger {
	service := path.Dir(fullMethodString)[1:]
	method := path.Base(fullMethodString)

	zerologCtx := logger.With()
	zerologCtx = SystemField(zerologCtx)
	zerologCtx = ClientField(zerologCtx)
	return zerologCtx.
		Str("grpc.service", service).
		Str("grpc.method", method).
		Logger()
}
