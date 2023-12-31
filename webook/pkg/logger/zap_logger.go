package logger

import (
	"strings"

	"go.uber.org/zap"
)

type ZapLogger struct {
	logger      *zap.Logger
	atomicLevel zap.AtomicLevel
}

func NewZapLogger(logger *zap.Logger, atomicLevel zap.AtomicLevel) *ZapLogger {
	return &ZapLogger{logger: logger, atomicLevel: atomicLevel}
}

func (z *ZapLogger) Debug(msg string, args ...any) {
	z.logger.Debug(msg, z.toArgs(args)...)
}

func (z *ZapLogger) Info(msg string, args ...any) {
	z.logger.Info(msg, z.toArgs(args)...)
}

func (z *ZapLogger) Warn(msg string, args ...any) {
	z.logger.Warn(msg, z.toArgs(args)...)
}

func (z *ZapLogger) Error(msg string, args ...any) {
	z.logger.Error(msg, z.toArgs(args)...)
}

func (z *ZapLogger) Panic(msg string, args ...any) {
	z.logger.Panic(msg, z.toArgs(args)...)
}

func (z *ZapLogger) Fatal(msg string, args ...any) {
	z.logger.Fatal(msg, z.toArgs(args)...)
}

func (z *ZapLogger) toArgs(args []any) []zap.Field {
	res := make([]zap.Field, len(args))
	for i := range args {
		ar := args[i].(Field)
		res[i] = zap.Any(ar.Key, ar.Value)
	}

	return res
}

func (z *ZapLogger) SetLogLevel(level string) {
	z.atomicLevel.SetLevel(ToZapLevel(level).Level())
}

func ToZapLevel(level string) zap.AtomicLevel {
	switch strings.ToLower(level) {
	default:
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case "debug":
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		return zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "panic":
		return zap.NewAtomicLevelAt(zap.PanicLevel)
	case "fatal":
		return zap.NewAtomicLevelAt(zap.FatalLevel)
	}
}
