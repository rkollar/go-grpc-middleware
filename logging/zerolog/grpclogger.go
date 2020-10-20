// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_zerolog

import (
	"fmt"

	"github.com/rs/zerolog"
	"google.golang.org/grpc/grpclog"
)

// ReplaceGrpcLogger sets the given zerolog.Logger as a gRPC-level logger.
// This should be called *before* any other initialization, preferably from init() functions.
// Deprecated: use ReplaceGrpcLoggerV2.
func ReplaceGrpcLogger(logger zerolog.Logger) {
	zerologCtx := logger.With()
	zerologCtx = SystemField(zerologCtx)
	logger = zerologCtx.
		Bool("grpc_log", true).
		Logger()

	zgl := &zerologGrpcLogger{logger}
	grpclog.SetLogger(zgl)
}

type zerologGrpcLogger struct {
	logger zerolog.Logger
}

func (l *zerologGrpcLogger) Fatal(args ...interface{}) {
	l.logger.Fatal().Msg(fmt.Sprint(args...))
}

func (l *zerologGrpcLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal().Msgf(format, args...)
}

func (l *zerologGrpcLogger) Fatalln(args ...interface{}) {
	l.logger.Fatal().Msg(fmt.Sprint(args...))
}

func (l *zerologGrpcLogger) Print(args ...interface{}) {
	l.logger.Info().Msg(fmt.Sprint(args...))
}

func (l *zerologGrpcLogger) Printf(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

func (l *zerologGrpcLogger) Println(args ...interface{}) {
	l.logger.Info().Msg(fmt.Sprint(args...))
}

// ReplaceGrpcLoggerV2 replaces the grpc_log.LoggerV2 with the provided logger.
func ReplaceGrpcLoggerV2(logger zerolog.Logger) {
	ReplaceGrpcLoggerV2WithVerbosity(logger, 0)
}

// ReplaceGrpcLoggerV2WithVerbosity replaces the grpc_.LoggerV2 with the provided logger and verbosity.
func ReplaceGrpcLoggerV2WithVerbosity(logger zerolog.Logger, verbosity int) {
	logger = SystemField(logger.With()).Bool("grpc_log", true).Logger()
	zgl := &zerologGrpcLoggerV2{
		logger:    logger,
		verbosity: verbosity,
	}
	grpclog.SetLoggerV2(zgl)
}

type zerologGrpcLoggerV2 struct {
	logger    zerolog.Logger
	verbosity int
}

func (l *zerologGrpcLoggerV2) Info(args ...interface{}) {
	l.logger.Info().Msg(fmt.Sprint(args...))
}

func (l *zerologGrpcLoggerV2) Infoln(args ...interface{}) {
	l.logger.Info().Msg(fmt.Sprint(args...))
}

func (l *zerologGrpcLoggerV2) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

func (l *zerologGrpcLoggerV2) Warning(args ...interface{}) {
	l.logger.Warn().Msg(fmt.Sprint(args...))
}

func (l *zerologGrpcLoggerV2) Warningln(args ...interface{}) {
	l.logger.Warn().Msg(fmt.Sprint(args...))
}

func (l *zerologGrpcLoggerV2) Warningf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

func (l *zerologGrpcLoggerV2) Error(args ...interface{}) {
	l.logger.Error().Msg(fmt.Sprint(args...))
}

func (l *zerologGrpcLoggerV2) Errorln(args ...interface{}) {
	l.logger.Error().Msg(fmt.Sprint(args...))
}

func (l *zerologGrpcLoggerV2) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

func (l *zerologGrpcLoggerV2) Fatal(args ...interface{}) {
	l.logger.Fatal().Msg(fmt.Sprint(args...))
}

func (l *zerologGrpcLoggerV2) Fatalln(args ...interface{}) {
	l.logger.Fatal().Msg(fmt.Sprint(args...))
}

func (l *zerologGrpcLoggerV2) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal().Msgf(format, args...)
}

func (l *zerologGrpcLoggerV2) V(level int) bool {
	return level <= l.verbosity
}
