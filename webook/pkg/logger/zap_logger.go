package logger

import "go.uber.org/zap"

type ZapLogger struct {
	logger *zap.Logger
}

func NewZapLogger(logger *zap.Logger) *ZapLogger {
	return &ZapLogger{logger: logger}
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
