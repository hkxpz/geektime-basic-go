package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"geektime-basic-go/webook/ioc"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	initViper()
	initLogger()
	closeFunc := ioc.InitOTEL()
	initPrometheus()
	app := InitApp()

	for _, consumer := range app.consumers {
		if err := consumer.Start(); err != nil {
			panic(err)
		}
	}

	server := app.web
	server.GET("/PING", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "PONG")
	})

	log.Println(server.Run(":8080"))
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	closeFunc(ctx)
}

func initPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatalln(http.ListenAndServe(":8081", nil))
	}()
}

func initViper() {
	cfile := pflag.String("config", "/etc/webook/config.yaml", "配置文件路径")
	pflag.Parse()
	viper.SetConfigFile(*cfile)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

func initViperWithWatchConfig() {
	initViper()
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		log.Println(in.Op)
	})
}

func InitViperRemote() {
	remote := pflag.String("remote", "127.0.0.1", "配置中心地址")
	pflag.Parse()
	viper.SetConfigType("yaml")
	if err := viper.AddRemoteProvider("etcd3", *remote, "/webook"); err != nil {
		panic(err)
	}
	if err := viper.ReadRemoteConfig(); err != nil {
		panic(err)
	}
}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("初始化 logger 失败: %s", err))
	}
	zap.ReplaceGlobals(logger)
}
