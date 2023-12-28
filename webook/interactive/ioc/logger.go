package ioc

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
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
	zapLogger, err := zcfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}

	l := logger.NewZapLogger(zapLogger, zcfg.Level)
	viper.OnConfigChange(func(in fsnotify.Event) {
		level := viper.GetString("log.level")
		if level == "" {
			zapLogger.Error("重新加载日志级别失败")
			return
		}
		l.SetLogLevel(level)
	})

	return l
}

func InitGlobalsLogger() {
	l, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("初始化 logger 失败: %s", err))
	}
	zap.ReplaceGlobals(l)
}
