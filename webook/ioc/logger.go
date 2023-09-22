package ioc

import (
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"geektime-basic-go/webook/pkg/logger"
)

func InitLogger() logger.Logger {
	cfg := struct {
		Level    string `yaml:"level"`
		Encoding string `yaml:"encoding"`
	}{}

	if err := viper.UnmarshalKey("log", &cfg); err != nil {
		panic(err)
	}

	zcfg := zap.NewDevelopmentConfig()
	zcfg.Level = toZapLevel(cfg.Level)
	zcfg.Encoding = cfg.Encoding
	l, err := zcfg.Build()
	if err != nil {
		panic(err)
	}
	return logger.NewZapLogger(l)
}

func toZapLevel(level string) zap.AtomicLevel {
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
