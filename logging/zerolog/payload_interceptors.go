package grpc_zerolog

import (
	"bytes"
	"context"
	"fmt"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/rkollar/go-grpc-middleware/logging"
	"github.com/rkollar/go-grpc-middleware/logging/zerolog/ctxzerolog"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

var (
	// JsonPbMarshaller is the marshaller used for serializing protobuf messages.
	// If needed, this variable can be reassigned with a different marshaller with the same Marshal() signature.
	JsonPbMarshaller grpc_logging.JsonPbMarshaler = &jsonpb.Marshaler{}
)

// PayloadUnaryServerInterceptor returns a new unary server interceptors that logs the payloads of requests.
//
// This *only* works when placed *after* the `grpc_zerolog.UnaryServerInterceptor`. However, the logging can be done to a
// separate instance of the logger.
func PayloadUnaryServerInterceptor(logger zerolog.Logger, decider grpc_logging.ServerPayloadLoggingDecider) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !decider(ctx, info.FullMethod, info.Server) {
			return handler(ctx, req)
		}
		// Use the provided zerolog.Logger for logging but use the fields from context.
		fields := ctxzerolog.TagsToFields(ctx)
		zerologCtx := logger.With().Fields(fields)
		zerologCtx = addServerCallFields(info.FullMethod, zerologCtx)
		logProtoMessageAsJson(zerologCtx.Logger(), req, "grpc.request.content", "server request payload logged as grpc.request.content field")
		resp, err := handler(ctx, req)
		if err == nil {
			logProtoMessageAsJson(zerologCtx.Logger(), resp, "grpc.response.content", "server response payload logged as grpc.response.content field")
		}
		return resp, err
	}
}

// PayloadStreamServerInterceptor returns a new server server interceptors that logs the payloads of requests.
//
// This *only* works when placed *after* the `grpc_zerolog.StreamServerInterceptor`. However, the logging can be done to a
// separate instance of the logger.
func PayloadStreamServerInterceptor(logger zerolog.Logger, decider grpc_logging.ServerPayloadLoggingDecider) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !decider(stream.Context(), info.FullMethod, srv) {
			return handler(srv, stream)
		}
		fields := ctxzerolog.TagsToFields(stream.Context())
		zerologCtx := logger.With().Fields(fields)
		zerologCtx = addServerCallFields(info.FullMethod, zerologCtx)
		newStream := &loggingServerStream{ServerStream: stream, logger: zerologCtx.Logger()}
		return handler(srv, newStream)
	}
}

// PayloadUnaryClientInterceptor returns a new unary client interceptor that logs the payloads of requests and responses.
func PayloadUnaryClientInterceptor(logger zerolog.Logger, decider grpc_logging.ClientPayloadLoggingDecider) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if !decider(ctx, method) {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		logEntry := newClientLogger(ctx, method, logger)
		logProtoMessageAsJson(logEntry, req, "grpc.request.content", "client request payload logged as grpc.request.content")
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err == nil {
			logProtoMessageAsJson(logEntry, reply, "grpc.response.content", "client response payload logged as grpc.response.content")
		}
		return err
	}
}

// PayloadStreamClientInterceptor returns a new streaming client interceptor that logs the payloads of requests and responses.
func PayloadStreamClientInterceptor(logger zerolog.Logger, decider grpc_logging.ClientPayloadLoggingDecider) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if !decider(ctx, method) {
			return streamer(ctx, desc, cc, method, opts...)
		}
		logEntry := newClientLogger(ctx, method, logger)
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		newStream := &loggingClientStream{ClientStream: clientStream, logger: logEntry}
		return newStream, err
	}
}

type loggingClientStream struct {
	grpc.ClientStream
	logger zerolog.Logger
}

func (l *loggingClientStream) SendMsg(m interface{}) error {
	err := l.ClientStream.SendMsg(m)
	if err == nil {
		logProtoMessageAsJson(l.logger, m, "grpc.request.content", "server request payload logged as grpc.request.content field")
	}
	return err
}

func (l *loggingClientStream) RecvMsg(m interface{}) error {
	err := l.ClientStream.RecvMsg(m)
	if err == nil {
		logProtoMessageAsJson(l.logger, m, "grpc.response.content", "server response payload logged as grpc.response.content field")
	}
	return err
}

type loggingServerStream struct {
	grpc.ServerStream
	logger zerolog.Logger
}

func (l *loggingServerStream) SendMsg(m interface{}) error {
	err := l.ServerStream.SendMsg(m)
	if err == nil {
		logProtoMessageAsJson(l.logger, m, "grpc.response.content", "server response payload logged as grpc.response.content field")
	}
	return err
}

func (l *loggingServerStream) RecvMsg(m interface{}) error {
	err := l.ServerStream.RecvMsg(m)
	if err == nil {
		logProtoMessageAsJson(l.logger, m, "grpc.request.content", "server request payload logged as grpc.request.content field")
	}
	return err
}

func logProtoMessageAsJson(logger zerolog.Logger, pbMsg interface{}, key string, msg string) {
	if p, ok := pbMsg.(proto.Message); ok {
		logger.WithLevel(zerolog.InfoLevel).
			Object(key, &zerologpbObjectMarshaler{j: jsonpbObjectMarshaler{pb: p}}).
			Msg(msg)
	}
}

type zerologpbObjectMarshaler struct {
	j jsonpbObjectMarshaler
}

func (z *zerologpbObjectMarshaler) MarshalZerologObject(e *zerolog.Event) {
	e.Interface("msg", &z.j)
}

type jsonpbObjectMarshaler struct {
	pb proto.Message
}

func (j *jsonpbObjectMarshaler) MarshalJSON() ([]byte, error) {
	b := &bytes.Buffer{}
	if err := JsonPbMarshaller.Marshal(b, j.pb); err != nil {
		return nil, fmt.Errorf("jsonpb serializer failed: %v", err)
	}
	return b.Bytes(), nil
}
