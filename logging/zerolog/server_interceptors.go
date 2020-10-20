package grpc_zerolog

import (
	"context"
	"path"
	"time"

	grpc_middleware "github.com/rkollar/go-grpc-middleware"
	"github.com/rkollar/go-grpc-middleware/logging/zerolog/ctxzerolog"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

var (
	// SystemField is used in every log statement made through grpc_zerolog. Can be overwritten before any initialization code.
	SystemField = func(c zerolog.Context) zerolog.Context {
		return c.Str("system", "grpc")
	}

	// ServerField is used in every server-side log statement made through grpc_zerolog.Can be overwritten before initialization.
	ServerField = func(c zerolog.Context) zerolog.Context {
		return c.Str("span.kind", "server")
	}
)

// UnaryServerInterceptor returns a new unary server interceptors that adds zerolog.Logger to the context.
func UnaryServerInterceptor(logger zerolog.Logger, opts ...Option) grpc.UnaryServerInterceptor {
	o := evaluateServerOpt(opts)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()

		newCtx := newLoggerForCall(ctx, logger, info.FullMethod, startTime)

		resp, err := handler(newCtx, req)
		if !o.shouldLog(info.FullMethod, err) {
			return resp, err
		}
		code := o.codeFunc(err)
		level := o.levelFunc(code)
		duration := o.durationFunc(time.Since(startTime))

		o.messageFunc(newCtx, "finished unary call with code "+code.String(), level, code, err, duration)
		return resp, err
	}
}

// StreamServerInterceptor returns a new streaming server interceptor that adds zerolog.Logger to the context.
func StreamServerInterceptor(logger zerolog.Logger, opts ...Option) grpc.StreamServerInterceptor {
	o := evaluateServerOpt(opts)
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		startTime := time.Now()
		newCtx := newLoggerForCall(stream.Context(), logger, info.FullMethod, startTime)
		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = newCtx

		err := handler(srv, wrapped)
		if !o.shouldLog(info.FullMethod, err) {
			return err
		}
		code := o.codeFunc(err)
		level := o.levelFunc(code)
		duration := o.durationFunc(time.Since(startTime))

		o.messageFunc(newCtx, "finished streaming call with code "+code.String(), level, code, err, duration)
		return err
	}
}

func addServerCallFields(fullMethodString string, zerologCtx zerolog.Context) zerolog.Context {
	service := path.Dir(fullMethodString)[1:]
	method := path.Base(fullMethodString)
	zerologCtx = SystemField(zerologCtx)
	zerologCtx = ServerField(zerologCtx)
	return zerologCtx.
		Str("grpc.service", service).
		Str("grpc.method", method)
}

func newLoggerForCall(ctx context.Context, logger zerolog.Logger, fullMethodString string, start time.Time) context.Context {
	zerologCtx := logger.With()
	zerologCtx = zerologCtx.Str("grpc.start_time", start.Format(time.RFC3339))
	if d, ok := ctx.Deadline(); ok {
		zerologCtx = zerologCtx.Str("grpc.request.deadline", d.Format(time.RFC3339))
	}

	zerologCtx = addServerCallFields(fullMethodString, zerologCtx)
	return ctxzerolog.ToContext(ctx, zerologCtx.Logger())
}
