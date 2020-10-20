package grpc_zerolog_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/rkollar/go-grpc-middleware/logging/zerolog/ctxzerolog"
	grpc_ctxtags "github.com/rkollar/go-grpc-middleware/tags"
	grpc_testing "github.com/rkollar/go-grpc-middleware/testing"
	pb_testproto "github.com/rkollar/go-grpc-middleware/testing/testproto"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
)

var (
	goodPing = &pb_testproto.PingRequest{Value: "something", SleepTimeMs: 9999}
)

type loggingPingService struct {
	pb_testproto.TestServiceServer
}

func (s *loggingPingService) Ping(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
	grpc_ctxtags.Extract(ctx).Set("custom_tags.string", "something").Set("custom_tags.int", 1337)
	ctxzerolog.AddFields(ctx, map[string]interface{}{"custom_field": "custom_value"})
	l := ctxzerolog.Extract(ctx)
	l.Info().Msg("some ping")
	return s.TestServiceServer.Ping(ctx, ping)
}

func (s *loggingPingService) PingError(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.Empty, error) {
	return s.TestServiceServer.PingError(ctx, ping)
}

func (s *loggingPingService) PingList(ping *pb_testproto.PingRequest, stream pb_testproto.TestService_PingListServer) error {
	grpc_ctxtags.Extract(stream.Context()).Set("custom_tags.string", "something").Set("custom_tags.int", 1337)
	l := ctxzerolog.Extract(stream.Context())
	l.Info().Msg("some pinglist")
	return s.TestServiceServer.PingList(ping, stream)
}

func (s *loggingPingService) PingEmpty(ctx context.Context, empty *pb_testproto.Empty) (*pb_testproto.PingResponse, error) {
	return s.TestServiceServer.PingEmpty(ctx, empty)
}

func newBaseZerologSuite(t *testing.T) *zerologBaseSuite {
	buffer := &bytes.Buffer{}
	mutexBuffer := grpc_testing.NewMutexReadWriter(buffer)

	zerolog.TimestampFieldName = "ts"
	zerolog.LevelFieldName = "level"
	zerolog.MessageFieldName = "msg"
	zerolog.ErrorFieldName = "error"
	zerolog.CallerFieldName = "caller"
	zerolog.ErrorStackFieldName = "stacktrace"
	zerolog.TimeFieldFormat = time.RFC3339

	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	var logger = zerolog.New(mutexBuffer).With().Timestamp().Logger()
	s := &zerologBaseSuite{
		log:         logger,
		buffer:      buffer,
		mutexBuffer: mutexBuffer,
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			TestService: &loggingPingService{&grpc_testing.TestPingService{T: t}},
		},
	}
	return s
}

type zerologBaseSuite struct {
	*grpc_testing.InterceptorTestSuite
	mutexBuffer *grpc_testing.MutexReadWriter
	buffer      *bytes.Buffer
	log         zerolog.Logger
}

func (s *zerologBaseSuite) SetupTest() {
	s.mutexBuffer.Lock()
	s.buffer.Reset()
	s.mutexBuffer.Unlock()
}

func (s *zerologBaseSuite) getOutputJSONs() []map[string]interface{} {
	ret := make([]map[string]interface{}, 0)
	dec := json.NewDecoder(s.mutexBuffer)

	for {
		var val map[string]interface{}
		err := dec.Decode(&val)
		if err == io.EOF {
			break
		}
		if err != nil {
			s.T().Fatalf("failed decoding output from zerolog JSON: %v", err)
		}

		ret = append(ret, val)
	}

	return ret
}

func StubMessageProducer(ctx context.Context, msg string, level zerolog.Level, code codes.Code, err error, duration map[string]interface{}) {
	// re-extract logger from ctx, as it may have extra fields that changed in the holder.
	l := ctxzerolog.Extract(ctx)
	l.WithLevel(level).
		Fields(duration).
		Err(err).
		Str("grpc.code", code.String()).
		Msg("custom message")
}
