package ioc

import (
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"geektime-basic-go/webook/pkg/logger"
)

func InitZapLogger() logger.Logger {
	cfg := struct {
		Level    string `yaml:"level"`
		Encoding string `yaml:"encoding"`
	}{
		Encoding: "console",
	}

	if err := viper.UnmarshalKey("log", &cfg); err != nil {
		panic(err)
	}

	zcfg := zap.NewDevelopmentConfig()
	zcfg.Level = logger.ToZapLevel(cfg.Level)
	zcfg.Encoding = cfg.Encoding
	l, err := zcfg.Build()
	if err != nil {
		panic(err)
	}
	return logger.NewZapLogger(l)
}

func InitGlobalsLogger() {
	l, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("初始化 logger 失败: %s", err))
	}
	zap.ReplaceGlobals(l)
}
